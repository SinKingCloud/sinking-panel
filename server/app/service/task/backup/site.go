package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"server/app/model"
	"sort"
	"strings"
)

func (s *Manager) parseSiteOptions(content json.RawMessage) (*SiteOptions, error) {
	options := &SiteOptions{
		AllSites: true,
		SiteIds:  []int64{},
	}
	if len(content) > 0 {
		if err := json.Unmarshal(content, &options); err != nil {
			return nil, errors.New("网站备份参数必须是 JSON 对象")
		}
	}
	if options == nil {
		return nil, errors.New("网站备份参数必须是 JSON 对象")
	}
	if options.AllSites {
		options.SiteIds = []int64{}
		return options, nil
	}
	if len(options.SiteIds) == 0 || len(options.SiteIds) > 1000 {
		return nil, errors.New("请选择 1 到 1000 个网站")
	}
	seen := make(map[int64]bool, len(options.SiteIds))
	for _, id := range options.SiteIds {
		if id <= 0 || seen[id] {
			return nil, errors.New("备份网站 ID 必须为正数且不能重复")
		}
		seen[id] = true
	}
	sort.Slice(options.SiteIds, func(i, j int) bool { return options.SiteIds[i] < options.SiteIds[j] })
	return options, nil
}

// runSites 查询网站并逐个归档，单个网站失败不影响其他网站。
// 使用任务上下文取消执行，不依赖网站 HTTP 服务的启停状态。
func (s *Manager) runSites(ctx context.Context, id int64, config Config, writeLog func(string)) (err error) {
	defer func() {
		if err == nil {
			return
		}
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			writeLog("备份计划已取消")
		} else {
			writeLog("备份计划失败: " + err.Error())
		}
	}()
	if id <= 0 {
		return errors.New("备份计划 ID 不合法")
	}
	if s.repositorySite == nil {
		return errors.New("网站仓储未初始化")
	}
	if s.repositorySiteBackup == nil {
		return errors.New("网站备份仓储未初始化")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if _, running := s.siteRuns.LoadOrStore(id, struct{}{}); running {
		return errors.New("该备份计划正在执行")
	}
	defer s.siteRuns.Delete(id)
	options, err := s.parseSiteOptions(config.Options)
	if err != nil {
		return err
	}
	writeLog("开始执行备份计划")
	items, err := s.repositorySite.SelectAll()
	if err != nil {
		return err
	}
	selected := make(map[int64]bool, len(options.SiteIds))
	for _, id := range options.SiteIds {
		selected[id] = true
	}
	sites := make([]*model.Site, 0, len(items))
	failed, completed, skipped := 0, 0, 0
	for _, site := range items {
		if err = ctx.Err(); err != nil {
			return err
		}
		if !options.AllSites && !selected[site.Id] {
			continue
		}
		delete(selected, site.Id)
		if strings.TrimSpace(site.Root) == "" {
			skipped++
			writeLog(fmt.Sprintf("跳过网站 #%d：未配置根目录", site.Id))
			continue
		}
		info, statErr := os.Stat(site.Root)
		if os.IsNotExist(statErr) || (statErr == nil && !info.IsDir()) {
			skipped++
			writeLog(fmt.Sprintf("跳过网站 #%d：根目录不存在或不是目录", site.Id))
			continue
		}
		if statErr != nil {
			failed++
			writeLog(fmt.Sprintf("网站 #%d 检查根目录失败: %v", site.Id, statErr))
			continue
		}
		sites = append(sites, site)
	}
	for _, id := range options.SiteIds {
		if selected[id] {
			failed++
			writeLog(fmt.Sprintf("网站 #%d 已不存在", id))
		}
	}
	sort.Slice(sites, func(i, j int) bool { return sites[i].Id < sites[j].Id })
	for _, site := range sites {
		if err = ctx.Err(); err != nil {
			return err
		}
		writeLog(fmt.Sprintf("开始备份网站 #%d %s", site.Id, site.Name))
		lastProgress, lastMessage := -1, ""
		record, runErr := s.SaveSite(ctx, site.Id, id, config, 0, func(progress float64, text string) {
			percent := int(progress)
			if percent != lastProgress || text != lastMessage {
				writeLog(fmt.Sprintf("%d%% %s", percent, text))
				lastProgress, lastMessage = percent, text
			}
		})
		if record != nil && runErr != nil && ctx.Err() == nil {
			writeLog("备份已保留：" + runErr.Error())
			// 存储不可用或清理失败不影响已生成的备份；有效存储上传失败仍计为失败。
			var data SiteData
			if config.Storage != "oss" || s.ossService == nil || s.ossService.Check(config.OssId) != nil || json.Unmarshal([]byte(record.Data), &data) == nil && data.OssPath != "" {
				runErr = nil
			}
		}
		if runErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if errors.Is(runErr, os.ErrNotExist) {
				info, statErr := os.Stat(site.Root)
				if os.IsNotExist(statErr) || (statErr == nil && !info.IsDir()) {
					skipped++
					writeLog(fmt.Sprintf("跳过网站 #%d：根目录已不存在或不是目录", site.Id))
					continue
				}
			}
			failed++
			writeLog(fmt.Sprintf("网站 #%d 备份失败: %v", site.Id, runErr))
			continue
		}
		completed++
		writeLog("100% 备份完成：" + record.Name)
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	message := fmt.Sprintf("备份结束，成功 %d 个，失败 %d 个，跳过 %d 个", completed, failed, skipped)
	writeLog(message)
	if failed > 0 {
		return errors.New(message)
	}
	writeLog("备份计划执行完成")
	return nil
}
