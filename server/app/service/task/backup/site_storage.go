package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"server/app/model"
	"server/app/util/archive"
	utilOss "server/app/util/oss"
	"server/app/util/str"
	"strconv"
	"strings"
	"time"
)

// SaveSite 同步归档、上传并记录网站备份；上传失败时仍保存本地副本和记录。
// taskId 为 0 时属于网站备份，否则仅清理同一计划任务生成的副本。
func (s *Manager) SaveSite(ctx context.Context, siteId, taskId int64, config Config, retries int, progress func(float64, string)) (result *model.SiteBackup, err error) {
	if siteId <= 0 || taskId < 0 {
		return nil, errors.New("备份网站或计划任务 ID 不合法")
	}
	if retries < 0 || retries > 10 {
		return nil, errors.New("重试次数必须为 0–10")
	}
	if s.repositorySite == nil || s.repositorySiteBackup == nil {
		return nil, errors.New("网站或备份仓储未初始化")
	}
	config, err = config.Validate()
	if err != nil {
		return nil, err
	}
	if progress == nil {
		progress = func(float64, string) {}
	}
	select {
	case s.gate <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.gate }()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	site, err := s.repositorySite.FindById(siteId)
	if err != nil {
		return nil, err
	}
	if site == nil {
		return nil, errors.New("网站不存在")
	}
	root, err := filepath.Abs(site.Root)
	if err != nil || strings.TrimSpace(site.Root) == "" {
		return nil, errors.New("网站没有可备份的目录")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	id := str.GetSnowWorkIns().GetId()
	name := time.Now().Format("20060102_150405") + "_" + strconv.FormatInt(id, 10) + ".tar.gz"
	localPath := filepath.Join(strconv.FormatInt(siteId, 10), name)
	destination := filepath.Join(s.siteRoot, localPath)
	if err = os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return nil, err
	}
	temporary := destination + ".pending"
	defer os.Remove(temporary)
	for attempt := 0; attempt <= retries; attempt++ {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		progress(0, "正在压缩网站文件")
		err = s.packBackup(ctx, site, temporary, func(value float64, _ string) {
			progress(value*0.6, "正在压缩网站文件")
		})
		if err == nil || ctx.Err() != nil {
			break
		}
		if attempt < retries {
			progress(0, fmt.Sprintf("压缩失败，开始第 %d 次重试", attempt+1))
		}
	}
	if err != nil {
		return nil, err
	}
	current, err := s.repositorySite.FindById(siteId)
	if err != nil {
		return nil, err
	}
	if current == nil || current.Root != site.Root {
		return nil, errors.New("网站已删除或根目录已修改，已取消保存备份")
	}
	progress(60, "正在校验备份文件")
	file, err := os.OpenFile(temporary, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	buffer := make([]byte, 64*1024)
	for {
		if err = ctx.Err(); err != nil {
			break
		}
		n, readErr := file.Read(buffer)
		if n > 0 {
			_, _ = hash.Write(buffer[:n])
		}
		if readErr != nil {
			if readErr != io.EOF {
				err = readErr
			}
			break
		}
	}
	info, statErr := file.Stat()
	err = errors.Join(err, statErr, file.Sync(), file.Close())
	if err != nil {
		return nil, err
	}
	if err = os.Rename(temporary, destination); err != nil {
		return nil, err
	}
	var client utilOss.Interface
	remotePath, saved := "", false
	defer func() {
		if saved {
			return
		}
		cleanupErr := os.Remove(destination)
		if errors.Is(cleanupErr, os.ErrNotExist) {
			cleanupErr = nil
		}
		if client != nil && remotePath != "" {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			if removeErr := client.DeleteFile(cleanupCtx, remotePath); !errors.Is(removeErr, utilOss.ErrNotFound) {
				cleanupErr = errors.Join(cleanupErr, removeErr)
			}
		}
		err = errors.Join(err, cleanupErr)
	}()
	data := SiteData{
		SiteName:  site.Name,
		TaskId:    taskId,
		Root:      root,
		LocalPath: filepath.ToSlash(localPath),
		Checksum:  hex.EncodeToString(hash.Sum(nil)),
	}
	record := &model.SiteBackup{
		Id:     id,
		SiteId: siteId,
		Name:   name,
		Size:   info.Size(),
	}
	var uploadErr error
	uploadOss := config.Storage == "oss"
	if uploadOss {
		progress(60, "正在上传云端副本")
		if s.ossService == nil {
			uploadErr = errors.New("云端存储服务未初始化")
		} else {
			client, uploadErr = s.ossService.Client(config.OssId)
			if uploadErr == nil {
				key := path.Join(config.Prefix, "site_backups", strconv.FormatInt(siteId, 10), name)
				for attempt := 0; attempt <= retries; attempt++ {
					if uploadErr = ctx.Err(); uploadErr != nil {
						break
					}
					var source *os.File
					source, uploadErr = os.Open(destination)
					if uploadErr != nil {
						break
					}
					remotePath = key
					_, uploadErr = client.Upload(ctx, key, source, record.Size, utilOss.UploadOptions{
						PutOptions: utilOss.PutOptions{ContentType: "application/gzip"},
						Progress: func(current, total int64) {
							if total > 0 {
								progress(60+float64(current)/float64(total)*35, "正在上传云端副本")
							}
						},
					})
					_ = source.Close()
					if uploadErr == nil {
						var remote *utilOss.File
						remote, uploadErr = client.StatFile(ctx, key)
						if uploadErr == nil && (remote == nil || remote.Size != record.Size) {
							uploadErr = errors.New("云端备份大小不一致")
						}
					}
					if uploadErr == nil {
						data.OssId, data.OssPath = config.OssId, key
						break
					}
					if attempt < retries {
						progress(60, fmt.Sprintf("上传失败，开始第 %d 次重试", attempt+1))
					}
				}
			}
		}
	}
	if uploadErr != nil && client != nil && remotePath != "" {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		removeErr := client.DeleteFile(cleanupCtx, remotePath)
		cancel()
		if removeErr == nil || errors.Is(removeErr, utilOss.ErrNotFound) {
			remotePath = ""
		} else {
			uploadErr = errors.Join(uploadErr, removeErr)
		}
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	// 对外发布一条完整记录，避免上传期间删除尚未完成的备份。
	content, _ := json.Marshal(data)
	record.Data = string(content)
	if err = s.repositorySiteBackup.Create(record); err != nil {
		return nil, err
	}
	saved = true
	localKeep := config.LocalKeep
	if uploadErr != nil {
		// 云端不可用时必须保留本次本地文件，即便配置为不保留本地副本。
		localKeep = max(localKeep, 1)
	}
	progress(95, "正在清理过期副本")
	cleanupErr := s.retainBackups(ctx, record, taskId, localKeep, config.OssKeep, uploadOss)
	if uploadErr != nil {
		uploadErr = fmt.Errorf("云端上传失败，本地备份已保留：%w", uploadErr)
	}
	return record, errors.Join(uploadErr, cleanupErr)
}

// retainBackups 按来源分别保留本地、云端副本，网站和不同计划任务互不清理。
func (s *Manager) retainBackups(ctx context.Context, published *model.SiteBackup, taskId int64, localKeep, ossKeep int, uploadOss bool) error {
	records, err := s.repositorySiteBackup.SelectBySiteId(published.SiteId)
	if err != nil {
		return err
	}
	// 校时可能使新记录排在旧记录之后，本次生成的备份始终优先保留。
	found := false
	for index, record := range records {
		if record.Id == published.Id {
			copy(records[1:index+1], records[:index])
			records[0] = published
			found = true
			break
		}
	}
	if !found {
		return errors.New("新备份未出现在记录列表中，已保留旧副本")
	}
	local, remote := 0, 0
	var failures []error
	for _, record := range records {
		if err = ctx.Err(); err != nil {
			return err
		}
		var data SiteData
		if err = json.Unmarshal([]byte(record.Data), &data); err != nil {
			failures = append(failures, fmt.Errorf("备份 %s 信息损坏：%w", record.Name, err))
			continue
		}
		if data.TaskId != taskId {
			continue
		}
		removeLocal, removeRemote := false, false
		if data.LocalPath != "" {
			local++
			removeLocal = local > localKeep
		}
		if data.OssPath != "" {
			remote++
			removeRemote = uploadOss && remote > ossKeep
		}
		if removeLocal || removeRemote {
			if err = s.DeleteSite(ctx, record, &data, removeLocal, removeRemote); err != nil {
				failures = append(failures, fmt.Errorf("清理备份 %s 失败：%w", record.Name, err))
			}
		}
	}
	return errors.Join(failures...)
}

// DeleteSite 清理指定副本后更新记录；剩余副本保留以便再次操作。
// 副本信息按原值更新，避免并行清理覆盖已删除的副本。
func (s *Manager) DeleteSite(ctx context.Context, record *model.SiteBackup, data *SiteData, local, remote bool) error {
	if s.repositorySiteBackup == nil {
		return errors.New("网站备份仓储未初始化")
	}
	if record == nil || record.Id <= 0 || data == nil {
		return errors.New("网站备份信息不合法")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	removedLocal, removedRemote := false, false
	save := func() error {
		for attempt := 0; attempt < 3; attempt++ {
			content, _ := json.Marshal(data)
			if data.LocalPath == "" && data.OssPath == "" {
				if err := s.repositorySiteBackup.DeleteByIds([]int64{record.Id}); err != nil {
					return err
				}
				record.Data = string(content)
				return nil
			}
			if string(content) == record.Data {
				return nil
			}
			updated, err := s.repositorySiteBackup.UpdateData(record.Id, record.Data, string(content))
			if err != nil {
				return err
			}
			if updated > 0 {
				record.Data = string(content)
				return nil
			}
			current, err := s.repositorySiteBackup.FindById(record.Id)
			if err != nil {
				return err
			}
			if current == nil {
				data.LocalPath, data.OssId, data.OssPath = "", 0, ""
				content, _ = json.Marshal(data)
				record.Data = string(content)
				return nil
			}
			var currentData SiteData
			if err = json.Unmarshal([]byte(current.Data), &currentData); err != nil {
				return err
			}
			*data = currentData
			record.Data = current.Data
			if removedLocal {
				data.LocalPath = ""
			}
			if removedRemote {
				data.OssId, data.OssPath = 0, ""
			}
		}
		return errors.New("备份副本信息已变更，请重试")
	}
	if local && data.LocalPath != "" {
		if !filepath.IsLocal(data.LocalPath) {
			return errors.New("备份文件路径不合法")
		}
		root, err := os.OpenRoot(s.siteRoot)
		if err == nil {
			err = root.Remove(filepath.FromSlash(data.LocalPath))
			_ = root.Close()
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		data.LocalPath = ""
		removedLocal = true
		// 删除云端失败时仍需准确保留剩余副本；仅删本地时由末尾一次落库。
		if remote && data.OssPath != "" {
			if err = save(); err != nil {
				return err
			}
		}
	}
	if remote && data.OssPath != "" {
		if s.ossService == nil {
			return errors.New("云端存储服务未初始化")
		}
		if data.OssId <= 0 {
			return errors.New("备份没有云端副本")
		}
		client, err := s.ossService.Client(data.OssId)
		if err != nil {
			return err
		}
		if err = client.DeleteFile(ctx, data.OssPath); err != nil && !errors.Is(err, utilOss.ErrNotFound) {
			return err
		}
		data.OssId, data.OssPath = 0, ""
		removedRemote = true
	}
	return save()
}

// packBackup 保留网站目录名，压缩前检查可恢复性并防止递归备份。
func (s *Manager) packBackup(ctx context.Context, site *model.Site, destination string, progress func(float64, string)) error {
	if site == nil || strings.TrimSpace(site.Root) == "" {
		return errors.New("网站未配置根目录")
	}
	root, err := filepath.Abs(site.Root)
	if err != nil {
		return fmt.Errorf("解析网站根目录失败: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("读取网站根目录失败: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("网站根目录必须是实际目录，不能是符号链接")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("解析网站根目录失败: %w", err)
	}
	if filepath.Dir(root) == root {
		return errors.New("不能备份整个磁盘或文件系统根目录")
	}
	for parent := s.siteRoot; ; parent = filepath.Dir(parent) {
		resolved, resolveErr := filepath.EvalSymlinks(parent)
		if resolveErr == nil {
			relative, relativeErr := filepath.Rel(parent, s.siteRoot)
			if relativeErr != nil {
				return relativeErr
			}
			backupRoot := filepath.Join(resolved, relative)
			for _, paths := range [][2]string{{root, backupRoot}, {backupRoot, root}} {
				relative, relativeErr = filepath.Rel(paths[0], paths[1])
				if relativeErr == nil && (relative == "." || filepath.IsLocal(relative)) {
					return errors.New("网站根目录不能包含备份目录，也不能位于备份目录中")
				}
			}
			break
		}
		if !os.IsNotExist(resolveErr) || filepath.Dir(parent) == parent {
			return fmt.Errorf("解析备份目录失败: %w", resolveErr)
		}
	}
	source, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("打开网站根目录失败: %w", err)
	}
	defer source.Close()
	err = fs.WalkDir(source.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if info.IsDir() || info.Mode().IsRegular() {
			return nil
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("网站包含不支持备份的特殊文件: %s", name)
		}
		link, linkErr := source.Readlink(name)
		if linkErr != nil {
			return linkErr
		}
		if link == "" || filepath.IsAbs(link) || filepath.VolumeName(link) != "" || !filepath.IsLocal(filepath.Join(filepath.Dir(name), link)) {
			return fmt.Errorf("网站包含外部或绝对路径的符号链接: %s", name)
		}
		if _, statErr := source.Stat(name); statErr != nil && !os.IsNotExist(statErr) {
			return fmt.Errorf("网站符号链接无法安全解析 %s: %w", name, statErr)
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = archive.Compress.CompressFiles(ctx, []string{root}, destination, archive.FormatTgz, func(current, total int64, name string, files, index int64) bool {
		if progress != nil {
			percent := float64(0)
			if total > 0 {
				percent = min(100, float64(current)*100/float64(total))
			} else if files > 0 {
				percent = min(100, float64(index)*100/float64(files))
			}
			progress(percent, name)
		}
		return ctx.Err() == nil
	})
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("压缩网站目录失败: %w", err)
	}
	current, err := os.Stat(root)
	if err != nil || !os.SameFile(info, current) {
		return errors.New("网站根目录在备份期间发生变动，已取消保存备份")
	}
	return nil
}
