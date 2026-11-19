package site

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"server/app/constant"
	"server/app/enum/site_backup_status"
	"server/app/enum/system_task_status"
	"server/app/enum/task_status"
	"server/app/model"
	repositorySiteBackup "server/app/repository/site_backup"
	repositoryTask "server/app/repository/task"
	taskBackup "server/app/service/task/backup"
	"server/app/util/archive"
	utilOss "server/app/util/oss"
	"server/app/util/page"
	"server/app/util/str"

	"github.com/robfig/cron/v3"
)

// GetBackupSettings 获取网站备份策略，默认开启，配置不与计划任务的备份共享。
func (s *service) GetBackupSettings() (map[string]string, error) {
	if s.configService == nil {
		return nil, errors.New("配置服务未初始化")
	}
	configs := map[string]string{
		constant.SiteBackupEnabled:   "true",
		constant.SiteBackupSpec:      "0 0 2 * * *",
		constant.SiteBackupRetries:   "0",
		constant.SiteBackupLocalKeep: "3",
		constant.SiteBackupUploadOss: "false",
		constant.SiteBackupOssId:     "0",
		constant.SiteBackupOssKeep:   "3",
		constant.SiteBackupPrefix:    "/backups",
	}
	stored := s.configService.Group(constant.SiteBackupGroup)
	for key := range configs {
		if value, exists := stored[key]; exists {
			configs[key] = value
		}
	}
	if err := s.validateBackupSettings(configs); err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *service) validateBackupSettings(configs map[string]string) error {
	enabled, err := strconv.ParseBool(configs[constant.SiteBackupEnabled])
	if err != nil {
		return errors.New("自动备份开关配置不合法")
	}
	spec := strings.TrimSpace(configs[constant.SiteBackupSpec])
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(spec); err != nil {
		return errors.New("备份执行周期不合法")
	}
	retries, err := strconv.Atoi(configs[constant.SiteBackupRetries])
	if err != nil || retries < 0 || retries > 10 {
		return errors.New("重试次数必须为 0–10")
	}
	uploadOss, err := strconv.ParseBool(configs[constant.SiteBackupUploadOss])
	if err != nil {
		return errors.New("备份上传云端开关配置不合法")
	}
	localKeep, err := strconv.Atoi(configs[constant.SiteBackupLocalKeep])
	if err != nil {
		return errors.New("本地副本数配置不合法")
	}
	ossId, err := strconv.ParseInt(configs[constant.SiteBackupOssId], 10, 64)
	if err != nil {
		return errors.New("备份存储实例配置不合法")
	}
	ossKeep, err := strconv.Atoi(configs[constant.SiteBackupOssKeep])
	if err != nil {
		return errors.New("云端保留副本数配置不合法")
	}
	prefix := strings.TrimSpace(configs[constant.SiteBackupPrefix])
	if uploadOss && prefix == "" {
		return errors.New("请设置云端备份目录")
	}
	config := taskBackup.Config{
		Storage:   "local",
		LocalKeep: localKeep,
		OssKeep:   ossKeep,
		OssId:     ossId,
		Prefix:    prefix,
	}
	if uploadOss {
		config.Storage = "oss"
	}
	if _, err = config.Validate(); err != nil {
		return err
	}
	configs[constant.SiteBackupEnabled] = strconv.FormatBool(enabled)
	configs[constant.SiteBackupSpec] = spec
	configs[constant.SiteBackupUploadOss] = strconv.FormatBool(uploadOss)
	configs[constant.SiteBackupPrefix] = prefix
	return nil
}

// initBackupTask 只借用内部调度器，不创建用户计划任务记录。
func (s *service) initBackupTask() error {
	if s.taskService == nil {
		return errors.New("任务服务未初始化")
	}
	settings, err := s.GetBackupSettings()
	if err != nil {
		return err
	}
	status := task_status.Stop
	if settings[constant.SiteBackupEnabled] == "true" {
		status = task_status.Running
	}
	content, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return s.taskService.Add(&model.Task{
		Id:     constant.TaskSiteBackup,
		Name:   "网站自动备份",
		Spec:   settings[constant.SiteBackupSpec],
		Script: string(content),
		Status: status,
	}, s.autoBackupSites)
}

// SetBackupSettings 保存配置并更新调度；失败时恢复原开关、周期及执行参数。
func (s *service) SetBackupSettings(configs map[string]string) error {
	settings := map[string]string{
		constant.SiteBackupEnabled:   configs[constant.SiteBackupEnabled],
		constant.SiteBackupSpec:      configs[constant.SiteBackupSpec],
		constant.SiteBackupRetries:   configs[constant.SiteBackupRetries],
		constant.SiteBackupLocalKeep: configs[constant.SiteBackupLocalKeep],
		constant.SiteBackupUploadOss: configs[constant.SiteBackupUploadOss],
		constant.SiteBackupOssId:     configs[constant.SiteBackupOssId],
		constant.SiteBackupOssKeep:   configs[constant.SiteBackupOssKeep],
		constant.SiteBackupPrefix:    configs[constant.SiteBackupPrefix],
	}
	if err := s.validateBackupSettings(settings); err != nil {
		return err
	}
	if settings[constant.SiteBackupEnabled] == "true" && settings[constant.SiteBackupUploadOss] == "true" {
		if s.ossService == nil {
			return errors.New("云端存储服务未初始化")
		}
		ossId, _ := strconv.ParseInt(settings[constant.SiteBackupOssId], 10, 64)
		if err := s.ossService.Check(ossId); err != nil {
			return err
		}
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.closed {
		return errors.New("网站服务已关闭")
	}
	previous, err := s.GetBackupSettings()
	if err != nil {
		return err
	}
	content, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	script := string(content)
	spec := settings[constant.SiteBackupSpec]
	status := task_status.Stop
	if settings[constant.SiteBackupEnabled] == "true" {
		status = task_status.Running
	}
	if s.backupTask && s.taskService != nil {
		if err = s.taskService.UpdateByIds([]int64{constant.TaskSiteBackup}, &repositoryTask.UpdateTask{
			Spec:   &spec,
			Script: &script,
			Status: &status,
		}); err != nil {
			return err
		}
	}
	if err = s.configService.Sets(settings); err != nil {
		if s.backupTask && s.taskService != nil {
			old, _ := json.Marshal(previous)
			oldScript := string(old)
			oldSpec := previous[constant.SiteBackupSpec]
			status = task_status.Stop
			if previous[constant.SiteBackupEnabled] == "true" {
				status = task_status.Running
			}
			err = errors.Join(err, s.taskService.UpdateByIds([]int64{constant.TaskSiteBackup}, &repositoryTask.UpdateTask{
				Spec:   &oldSpec,
				Script: &oldScript,
				Status: &status,
			}))
		}
		return err
	}
	return nil
}

// ReadBackupLog 复用内部自动备份任务日志，支持增量与历史读取。
func (s *service) ReadBackupLog(after, before int64, pageSize int) (map[string]interface{}, error) {
	if s.taskService == nil {
		return nil, errors.New("任务服务未初始化")
	}
	return s.taskService.ReadLog(constant.TaskSiteBackup, after, before, pageSize)
}

// ClearBackupLog 清理内部自动备份任务日志。
func (s *service) ClearBackupLog() error {
	if s.taskService == nil {
		return errors.New("任务服务未初始化")
	}
	return s.taskService.ClearLog(constant.TaskSiteBackup)
}

// SelectBackup 使用数据库分页；网站删除后仍显示备份时的网站名称。
func (s *service) SelectBackup(where *repositorySiteBackup.SelectSiteBackup, query *page.Query) (*page.Result[*repositorySiteBackup.SiteBackup], error) {
	result, err := s.repositorySiteBackup.Select(where, query)
	if err != nil {
		return nil, err
	}
	for _, record := range result.List {
		if record.SiteName == "" {
			var data taskBackup.SiteData
			if json.Unmarshal([]byte(record.Data), &data) == nil {
				record.SiteName = data.SiteName
			}
		}
	}
	return result, nil
}

func (s *service) autoBackupSites(ctx context.Context, _ int64, content string, writeLog func(string)) error {
	var settings map[string]string
	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		return errors.New("网站备份任务配置不合法")
	}
	if err := s.validateBackupSettings(settings); err != nil {
		return err
	}
	if settings[constant.SiteBackupEnabled] != "true" {
		return nil
	}
	if writeLog == nil {
		writeLog = func(string) {}
	}
	sites, err := s.repositorySite.SelectAutoBackup()
	if err != nil {
		return err
	}
	writeLog(fmt.Sprintf("开始网站自动备份，共 %d 个网站", len(sites)))
	var failures []error
	for _, site := range sites {
		if err := ctx.Err(); err != nil {
			return err
		}
		current, err := s.repositorySite.FindById(site.Id)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if current == nil || current.BackupStatus != site_backup_status.Enabled {
			continue
		}
		info, statErr := os.Stat(strings.TrimSpace(current.Root))
		if strings.TrimSpace(current.Root) == "" || errors.Is(statErr, os.ErrNotExist) || (statErr == nil && !info.IsDir()) {
			writeLog(fmt.Sprintf("跳过网站 #%d %s：没有网站目录", current.Id, current.Name))
			continue
		}
		if statErr != nil {
			failures = append(failures, statErr)
			writeLog(fmt.Sprintf("网站 #%d 读取目录失败：%v", current.Id, statErr))
			continue
		}
		writeLog(fmt.Sprintf("开始备份网站 #%d %s", current.Id, current.Name))
		lastMessage := ""
		record, err := s.createBackup(ctx, current.Id, settings, func(_ float64, text string) {
			if text != lastMessage {
				writeLog(text)
				lastMessage = text
			}
		})
		if err != nil {
			failures = append(failures, err)
			writeLog(fmt.Sprintf("网站 #%d 备份未完全完成：%v", current.Id, err))
		} else {
			writeLog("备份完成：" + record.Name)
		}
	}
	writeLog(fmt.Sprintf("网站备份检查结束，失败 %d 个", len(failures)))
	return errors.Join(failures...)
}

// CreateBackup 一次提交一个后台任务，按所选顺序备份网站并汇总执行结果。
func (s *service) CreateBackup(siteIds []int64) (string, error) {
	if len(siteIds) == 0 || len(siteIds) > 1000 {
		return "", errors.New("请选择 1–1000 个网站")
	}
	for _, siteId := range siteIds {
		if siteId <= 0 {
			return "", errors.New("网站ID不合法")
		}
	}
	settings, err := s.GetBackupSettings()
	if err != nil {
		return "", err
	}
	// 后台执行使用独立快照，避免调用方修改切片影响排队中的任务。
	siteIds = append([]int64(nil), siteIds...)
	id := "site_backup_" + strconv.FormatInt(str.GetSnowWorkIns().GetId(), 10)
	err = s.backupTasks.Create(id, fmt.Sprintf("备份网站：共 %d 个", len(siteIds)), map[string][]int64{"site_ids": siteIds}, func(ctx context.Context, _ interface{}, update func(int, float64, string)) {
		var failures []error
		for index, siteId := range siteIds {
			if ctx.Err() != nil {
				return
			}
			label := fmt.Sprintf("[%d/%d] 网站 #%d：", index+1, len(siteIds), siteId)
			update(system_task_status.Running, float64(index)*100/float64(len(siteIds)), label+"开始备份")
			_, err := s.createBackup(ctx, siteId, settings, func(progress float64, text string) {
				update(system_task_status.Running, (float64(index)*100+progress)/float64(len(siteIds)), label+text)
			})
			if ctx.Err() != nil {
				return
			}
			text := "备份完成"
			if err != nil {
				failures = append(failures, fmt.Errorf("网站 #%d：%w", siteId, err))
				text = "备份未完全完成：" + err.Error()
			}
			update(system_task_status.Running, float64(index+1)*100/float64(len(siteIds)), label+text)
		}
		text := fmt.Sprintf("网站备份结束：成功 %d 个，失败 %d 个", len(siteIds)-len(failures), len(failures))
		if err := errors.Join(failures...); err != nil {
			update(system_task_status.Failed, 100, text+"\n"+err.Error())
			return
		}
		update(system_task_status.Completed, 100, text)
	})
	return id, err
}

// createBackup 使用网站独立策略保存备份；计划任务自行保存，不依赖网站服务。
func (s *service) createBackup(ctx context.Context, siteId int64, settings map[string]string, progress func(float64, string)) (*model.SiteBackup, error) {
	ctx, release, err := s.operationContext(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	select {
	case s.backupGate <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.backupGate }()
	config := taskBackup.Config{
		Storage: "local",
		Prefix:  settings[constant.SiteBackupPrefix],
	}
	config.LocalKeep, _ = strconv.Atoi(settings[constant.SiteBackupLocalKeep])
	config.OssKeep, _ = strconv.Atoi(settings[constant.SiteBackupOssKeep])
	config.OssId, _ = strconv.ParseInt(settings[constant.SiteBackupOssId], 10, 64)
	if settings[constant.SiteBackupUploadOss] == "true" {
		config.Storage = "oss"
	}
	retries, _ := strconv.Atoi(settings[constant.SiteBackupRetries])
	return s.backup.SaveSite(ctx, siteId, 0, config, retries, progress)
}

func (s *service) findBackup(id int64) (*model.SiteBackup, *taskBackup.SiteData, error) {
	record, err := s.repositorySiteBackup.FindById(id)
	if err != nil {
		return nil, nil, err
	}
	if record == nil {
		return nil, nil, errors.New("备份不存在")
	}
	var data taskBackup.SiteData
	if err = json.Unmarshal([]byte(record.Data), &data); err != nil {
		return nil, nil, errors.New("备份信息损坏")
	}
	return record, &data, nil
}

func (s *service) backupOssClient(data *taskBackup.SiteData) (utilOss.Interface, error) {
	if s.ossService == nil {
		return nil, errors.New("云端存储服务未初始化")
	}
	if data.OssId <= 0 || data.OssPath == "" {
		return nil, errors.New("备份没有云端副本")
	}
	return s.ossService.Client(data.OssId)
}

func (s *service) openBackupFile(localPath string) (*os.File, error) {
	if !filepath.IsLocal(localPath) {
		return nil, errors.New("备份文件路径不合法")
	}
	root, err := os.OpenRoot(s.backupPath)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	file, err := root.Open(filepath.FromSlash(localPath))
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, errors.New("备份文件不可用")
	}
	return file, nil
}

// OpenBackup 默认优先本地；本地缺失时尝试云端，不将文件加载到内存。
func (s *service) OpenBackup(ctx context.Context, id int64, storage string) (io.ReadCloser, string, int64, error) {
	if storage != "" && storage != "local" && storage != "oss" {
		return nil, "", 0, errors.New("备份存储类型不合法")
	}
	record, data, err := s.findBackup(id)
	if err != nil {
		return nil, "", 0, err
	}
	if storage != "oss" && data.LocalPath != "" {
		file, openErr := s.openBackupFile(data.LocalPath)
		if openErr == nil {
			return file, record.Name, record.Size, nil
		}
		if storage == "local" || data.OssPath == "" {
			return nil, "", 0, openErr
		}
	}
	if storage == "local" {
		return nil, "", 0, errors.New("本地副本已清理")
	}
	client, err := s.backupOssClient(data)
	if err != nil {
		return nil, "", 0, err
	}
	result, err := client.GetFile(ctx, data.OssPath, utilOss.GetOptions{})
	if err != nil {
		return nil, "", 0, err
	}
	if result == nil || result.Body == nil {
		return nil, "", 0, errors.New("云端备份读取失败")
	}
	return result.Body, record.Name, record.Size, nil
}

// DownloadBackup 本地副本返回文件签名，云端副本返回存储厂商的签名 URL。
func (s *service) DownloadBackup(id int64, storage string) (string, error) {
	if storage != "" && storage != "local" && storage != "oss" {
		return "", errors.New("备份存储类型不合法")
	}
	_, data, err := s.findBackup(id)
	if err != nil {
		return "", err
	}
	if storage != "oss" && data.LocalPath != "" {
		file, openErr := s.openBackupFile(data.LocalPath)
		if openErr == nil {
			_ = file.Close()
			if s.fileService == nil {
				return "", errors.New("文件服务未初始化")
			}
			return s.fileService.CreatePreviewSign(filepath.Join(s.backupPath, filepath.FromSlash(data.LocalPath)), true, constant.CacheTimeWithFilePreview)
		}
		if storage == "local" || data.OssPath == "" {
			return "", openErr
		}
	}
	if storage == "local" {
		return "", errors.New("本地副本已清理")
	}
	client, err := s.backupOssClient(data)
	if err != nil {
		return "", err
	}
	signed, err := client.Sign(s.ctx, data.OssPath, "GET", 15*time.Minute, utilOss.SignOptions{
		ContentDisposition: mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(data.OssPath)}),
	})
	if err != nil {
		return "", err
	}
	return signed.URL, nil
}

// DeleteBackup 文件清理成功后再更新记录；部分失败保留剩余副本以便再次操作。
func (s *service) DeleteBackup(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("请选择要删除的备份")
	}
	ctx, release, err := s.operationContext(s.ctx)
	if err != nil {
		return err
	}
	defer release()
	select {
	case s.backupGate <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("备份或恢复任务正在执行，请稍后删除")
	}
	defer func() { <-s.backupGate }()
	records, err := s.repositorySiteBackup.SelectByIds(ids)
	if err != nil {
		return err
	}
	var failures []error
	for _, record := range records {
		if err = ctx.Err(); err != nil {
			return err
		}
		var data taskBackup.SiteData
		if err = json.Unmarshal([]byte(record.Data), &data); err == nil {
			err = s.backup.DeleteSite(ctx, record, &data, true, true)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("%s：%w", record.Name, err))
		}
	}
	return errors.Join(failures...)
}

// RestoreBackup 在后台流式校验和解压，再替换网站目录；旧目录保留到回收站。
func (s *service) RestoreBackup(id int64, siteId int64) (string, error) {
	record, _, err := s.findBackup(id)
	if err != nil {
		return "", err
	}
	if siteId == 0 {
		siteId = record.SiteId
	}
	site, err := s.repositorySite.FindById(siteId)
	if err != nil {
		return "", err
	}
	if site == nil || strings.TrimSpace(site.Root) == "" {
		return "", errors.New("恢复目标网站不存在或未配置网站目录")
	}
	taskId := "site_restore_" + strconv.FormatInt(str.GetSnowWorkIns().GetId(), 10)
	err = s.backupTasks.Create(taskId, "恢复网站："+site.Name, map[string]int64{"site_id": siteId, "backup_id": id}, func(ctx context.Context, _ interface{}, update func(int, float64, string)) {
		run := func() error {
			ctx, release, err := s.operationContext(ctx)
			if err != nil {
				return err
			}
			defer release()
			select {
			case s.backupGate <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			defer func() { <-s.backupGate }()
			_, data, err := s.findBackup(id)
			if err != nil {
				return err
			}
			reader, _, size, err := s.OpenBackup(ctx, id, "")
			if err != nil {
				return err
			}
			defer reader.Close()
			stop := context.AfterFunc(ctx, func() { _ = reader.Close() })
			defer stop()
			if err = os.MkdirAll(s.backupPath, 0700); err != nil {
				return err
			}
			file, err := os.CreateTemp(s.backupPath, ".restore-*.tar.gz")
			if err != nil {
				return err
			}
			defer os.Remove(file.Name())
			defer file.Close()
			hash := sha256.New()
			buffer := make([]byte, 64*1024)
			var received int64
			for {
				if err = ctx.Err(); err != nil {
					return err
				}
				n, readErr := reader.Read(buffer)
				if n > 0 {
					if _, err = file.Write(buffer[:n]); err != nil {
						return err
					}
					_, _ = hash.Write(buffer[:n])
					received += int64(n)
					if size > 0 {
						update(system_task_status.Running, float64(received)/float64(size)*35, "正在读取并校验备份")
					}
				}
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					return readErr
				}
			}
			if received != size || data.Checksum == "" || hex.EncodeToString(hash.Sum(nil)) != data.Checksum {
				return errors.New("备份文件大小或校验值不一致，已停止恢复")
			}
			if err = file.Close(); err != nil {
				return err
			}
			current, err := s.repositorySite.FindById(siteId)
			if err != nil {
				return err
			}
			if current == nil || current.Root != site.Root {
				return errors.New("恢复目标网站已删除或目录已修改，请重新操作")
			}
			return s.restoreBackupArchive(ctx, current, file.Name(), filepath.Base(data.Root), func(value float64, text string) {
				update(system_task_status.Running, 35+value*0.65, text)
			})
		}
		if err := run(); err != nil {
			update(system_task_status.Failed, 0, err.Error())
			return
		}
		update(system_task_status.Completed, 100, "网站文件已恢复，原目录已保留到回收站")
	})
	return taskId, err
}

// restoreBackupArchive 先在同盘校验和解压，再替换网站目录；旧文件成功进入回收站后才清理临时目录。
func (s *service) restoreBackupArchive(ctx context.Context, site *model.Site, archivePath, sourceRoot string, progress func(float64, string)) error {
	if site == nil || strings.TrimSpace(site.Root) == "" {
		return errors.New("网站未配置根目录")
	}
	if s.repositorySite == nil || s.recycleService == nil {
		return errors.New("网站仓储或回收站服务未初始化")
	}
	if sourceRoot == "." || !filepath.IsLocal(sourceRoot) || filepath.Base(sourceRoot) != sourceRoot {
		return errors.New("备份中的网站根目录名称无效")
	}
	if archive.Compress.GetFormatByExt(archivePath) != archive.FormatTgz {
		return errors.New("网站备份必须是 tar.gz 格式")
	}
	root, err := filepath.Abs(site.Root)
	if err != nil {
		return fmt.Errorf("解析网站根目录失败: %w", err)
	}
	previous, err := os.Lstat(root)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("检查网站根目录失败: %w", err)
	}
	if previous != nil && (!previous.IsDir() || previous.Mode()&os.ModeSymlink != 0) {
		return errors.New("网站根目录必须是实际目录，不能是符号链接")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(root))
	if err != nil {
		return fmt.Errorf("读取网站父目录失败: %w", err)
	}
	root = filepath.Join(parent, filepath.Base(root))
	contains := func(parent, child string) bool {
		relative, relativeErr := filepath.Rel(parent, child)
		return relativeErr == nil && (relative == "." || filepath.IsLocal(relative))
	}
	resolve := func(name string) (string, error) {
		name, resolveErr := filepath.Abs(name)
		if resolveErr != nil {
			return "", resolveErr
		}
		for parent := name; ; parent = filepath.Dir(parent) {
			resolved, resolveErr := filepath.EvalSymlinks(parent)
			if resolveErr == nil {
				relative, relativeErr := filepath.Rel(parent, name)
				return filepath.Join(resolved, relative), relativeErr
			}
			if !os.IsNotExist(resolveErr) || filepath.Dir(parent) == parent {
				return "", resolveErr
			}
		}
	}
	panelRoot, err := resolve(constant.BasePath)
	if err != nil {
		return fmt.Errorf("解析面板目录失败: %w", err)
	}
	managedRoot, err := resolve(constant.SiteRootPath)
	if err != nil {
		return fmt.Errorf("解析网站目录失败: %w", err)
	}
	backupRoot, err := resolve(s.backupPath)
	if err != nil {
		return fmt.Errorf("解析备份目录失败: %w", err)
	}
	if filepath.Dir(root) == root || contains(root, panelRoot) || contains(root, managedRoot) ||
		(contains(panelRoot, root) && !contains(managedRoot, root)) || contains(root, backupRoot) || contains(backupRoot, root) {
		return errors.New("网站根目录涉及系统根目录、面板或备份数据，不能自动恢复")
	}
	protected := []string{"/bin", "/boot", "/dev", "/etc", "/lib", "/lib64", "/proc", "/run", "/sbin", "/sys", "/usr", "/var/lib", "/var/log", "/var/run", "/var/spool"}
	shared := []string{"/data", "/home", "/media", "/mnt", "/opt", "/root", "/srv", "/tmp", "/var/tmp", "/var/www", "/www"}
	if runtime.GOOS == "windows" {
		protected = []string{os.Getenv("SystemRoot"), os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramData")}
		shared = []string{filepath.Join(filepath.VolumeName(root)+string(filepath.Separator), "Users")}
	} else if runtime.GOOS == "darwin" {
		protected = append(protected, "/Applications", "/Library", "/System", "/private/etc", "/private/var")
		shared = append(shared, "/Users", "/Volumes")
	}
	if home, homeErr := os.UserHomeDir(); homeErr == nil && home != "" {
		shared = append(shared, home)
	}
	for index, paths := range [][]string{protected, shared} {
		for _, name := range paths {
			if name == "" {
				continue
			}
			resolved, resolveErr := resolve(name)
			if resolveErr != nil {
				return fmt.Errorf("检查受保护目录失败: %w", resolveErr)
			}
			if contains(root, resolved) || (index == 0 && contains(resolved, root)) {
				return errors.New("不能恢复到系统目录、用户主目录或网站公共目录")
			}
		}
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// 完整读取压缩流以校验 CRC，且只接受同一个目录下的普通文件、目录和内部相对链接。
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("打开备份文件失败: %w", err)
	}
	rootFound := false
	err = func() error {
		defer file.Close()
		reader, readErr := gzip.NewReader(file)
		if readErr != nil {
			return readErr
		}
		defer reader.Close()
		contents := tar.NewReader(reader)
		buffer := make([]byte, 32*1024)
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			header, readErr := contents.Next()
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return readErr
			}
			name := filepath.Clean(filepath.FromSlash(header.Name))
			if !filepath.IsLocal(name) || !contains(sourceRoot, name) {
				return fmt.Errorf("归档包含网站目录之外的文件: %s", header.Name)
			}
			switch header.Typeflag {
			case tar.TypeDir:
				rootFound = rootFound || name == sourceRoot
			case tar.TypeReg:
				if name == sourceRoot {
					return errors.New("归档根节点必须是目录")
				}
			case tar.TypeSymlink:
				link := filepath.FromSlash(header.Linkname)
				if name == sourceRoot || filepath.IsAbs(link) || filepath.VolumeName(link) != "" || !contains(sourceRoot, filepath.Join(filepath.Dir(name), link)) {
					return fmt.Errorf("归档包含网站目录之外的符号链接: %s", header.Name)
				}
			default:
				return fmt.Errorf("备份包含不支持恢复的特殊文件: %s", header.Name)
			}
			for {
				if err := ctx.Err(); err != nil {
					return err
				}
				_, readErr = contents.Read(buffer)
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					return readErr
				}
			}
		}
		if !rootFound {
			return errors.New("备份缺少唯一的网站根目录")
		}
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			_, readErr = reader.Read(buffer)
			if readErr == io.EOF {
				return nil
			}
			if readErr != nil {
				return readErr
			}
		}
	}()
	if err != nil {
		return fmt.Errorf("校验网站备份失败: %w", err)
	}
	destination, err := os.OpenRoot(parent)
	if err != nil {
		return fmt.Errorf("打开网站父目录失败: %w", err)
	}
	defer destination.Close()
	temporary, err := os.MkdirTemp(parent, ".site-restore-")
	if err != nil {
		return fmt.Errorf("创建同盘恢复目录失败: %w", err)
	}
	keepPrevious := false
	defer func() {
		if !keepPrevious {
			_ = destination.RemoveAll(filepath.Base(temporary))
		}
	}()
	extracted := filepath.Join(temporary, "files")
	err = archive.Compress.Extract(ctx, archivePath, extracted, func(current, total int64, name string, files, index int64) bool {
		if progress != nil {
			percent := float64(10)
			if total > 0 {
				percent += min(75, float64(current)*75/float64(total))
			}
			progress(percent, name)
		}
		return ctx.Err() == nil
	}, true)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("解压网站备份失败: %w", err)
	}
	oldRoot := filepath.Join(filepath.Base(temporary), "previous")
	err = func() error {
		s.operationMu.Lock()
		defer s.operationMu.Unlock()
		if err := ctx.Err(); err != nil {
			return err
		}
		current, findErr := s.repositorySite.FindById(site.Id)
		if findErr != nil {
			return fmt.Errorf("核对网站信息失败: %w", findErr)
		}
		if current == nil || current.Root != site.Root {
			return errors.New("网站已删除或根目录已修改，已取消恢复")
		}
		resolved, resolveErr := resolve(site.Root)
		if resolveErr != nil || resolved != root {
			return errors.New("网站目录路径在恢复期间发生变动，已取消恢复")
		}
		actual, statErr := destination.Lstat(filepath.Base(root))
		if (previous == nil && !os.IsNotExist(statErr)) || (previous != nil && (statErr != nil || !os.SameFile(previous, actual))) {
			return errors.New("网站根目录在恢复期间被替换，已取消恢复")
		}
		all, selectErr := s.repositorySite.SelectAll()
		if selectErr != nil {
			return fmt.Errorf("检查网站目录占用失败: %w", selectErr)
		}
		for _, other := range all {
			if other.Id == site.Id || strings.TrimSpace(other.Root) == "" {
				continue
			}
			otherRoot, resolveErr := resolve(other.Root)
			if resolveErr != nil {
				return fmt.Errorf("检查网站[%s]目录失败: %w", other.Name, resolveErr)
			}
			if contains(root, otherRoot) || contains(otherRoot, root) {
				return fmt.Errorf("网站目录与网站[%s]共用或嵌套，不能直接覆盖恢复", other.Name)
			}
		}
		if progress != nil {
			progress(90, "正在替换网站目录")
		}
		if previous != nil {
			if err := destination.Rename(filepath.Base(root), oldRoot); err != nil {
				return fmt.Errorf("暂存原网站目录失败；请检查目录是否被占用或为独立挂载点，原文件未更改: %w", err)
			}
			keepPrevious = true
		}
		newRoot := filepath.Join(filepath.Base(temporary), "files", sourceRoot)
		if err := destination.Rename(newRoot, filepath.Base(root)); err != nil {
			if previous != nil {
				if rollbackErr := destination.Rename(oldRoot, filepath.Base(root)); rollbackErr != nil {
					return fmt.Errorf("替换网站目录失败: %v；回滚失败: %v；原文件保留在 %s", err, rollbackErr, filepath.Join(parent, oldRoot))
				}
				keepPrevious = false
			}
			return fmt.Errorf("替换网站目录失败，原目录已保留；请检查跨磁盘挂载或文件占用: %w", err)
		}
		return nil
	}()
	if err != nil {
		return err
	}
	if keepPrevious {
		if progress != nil {
			progress(95, "正在将原网站文件移至回收站")
		}
		if err = s.recycleService.Create(filepath.Join(parent, oldRoot)); err != nil {
			return fmt.Errorf("网站已恢复，但原文件移至回收站失败: %v；请保留并检查 %s", err, filepath.Join(parent, oldRoot))
		}
		keepPrevious = false
	}
	if progress != nil {
		progress(100, "网站恢复完成")
	}
	return nil
}
