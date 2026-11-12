package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/app/model"
	"sort"
	"strconv"
	"strings"
	"time"
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
	failed, completed := 0, 0
	for _, site := range items {
		if err = ctx.Err(); err != nil {
			return err
		}
		if !options.AllSites && !selected[site.Id] {
			continue
		}
		delete(selected, site.Id)
		if strings.TrimSpace(site.Root) == "" {
			if !options.AllSites {
				failed++
				writeLog(fmt.Sprintf("网站 #%d 未配置根目录", site.Id))
			}
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
		file, runErr := s.Save(ctx, Target{
			Key:         "site",
			ID:          strconv.FormatInt(site.Id, 10),
			PlanId:      id,
			Extension:   ".tar.gz",
			ContentType: "application/gzip",
		}, config, func(ctx context.Context, output io.Writer) error {
			return s.packSiteArchive(ctx, site, output)
		})
		if runErr != nil {
			if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
				return runErr
			}
			failed++
			writeLog(fmt.Sprintf("网站 #%d 备份失败: %v", site.Id, runErr))
			continue
		}
		completed++
		writeLog("备份完成: " + file.Name)
		if file.Warning != "" {
			writeLog(file.Warning)
		}
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	message := fmt.Sprintf("备份结束，成功 %d 个，失败 %d 个", completed, failed)
	writeLog(message)
	if failed > 0 {
		return errors.New(message)
	}
	writeLog("备份计划执行完成")
	return nil
}

// packSiteArchive 复用文件归档流程，保留网站清单并在发布前核对根目录。
func (s *Manager) packSiteArchive(ctx context.Context, site *model.Site, output io.Writer) (err error) {
	if strings.TrimSpace(site.Root) == "" {
		return errors.New("网站未配置根目录")
	}
	sourcePath, err := filepath.Abs(site.Root)
	if err != nil {
		return err
	}
	sourcePath, err = s.validateSourcePath(sourcePath, true)
	if err != nil {
		return err
	}
	source, err := os.OpenRoot(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	compressed := gzip.NewWriter(output)
	archive := tar.NewWriter(compressed)
	defer func() {
		err = errors.Join(err, archive.Close(), compressed.Close(), ctx.Err())
		if err != nil {
			return
		}
		current, findErr := s.repositorySite.FindById(site.Id)
		if findErr != nil {
			err = fmt.Errorf("核对网站信息失败: %w", findErr)
		} else if current == nil || current.Root != site.Root {
			err = errors.New("网站已删除或根目录已修改，已取消发布备份")
		}
	}()
	metadata, err := json.Marshal(struct {
		Version int         `json:"version"`
		Time    time.Time   `json:"time"`
		Site    *model.Site `json:"site"`
	}{
		Version: 1,
		Time:    time.Now().UTC(),
		Site:    site,
	})
	if err != nil {
		return err
	}
	if err = archive.WriteHeader(&tar.Header{
		Name:    "backup.json",
		Mode:    0600,
		Size:    int64(len(metadata)),
		ModTime: time.Now(),
	}); err != nil {
		return err
	}
	if _, err = archive.Write(metadata); err != nil {
		return err
	}
	return s.writePathEntry(ctx, source, ".", "files", archive, make([]byte, 32*1024))
}
