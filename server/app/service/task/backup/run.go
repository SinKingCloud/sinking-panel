package backup

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	serviceOss "server/app/service/oss"
	"server/app/util/oss"
	"strconv"
	"strings"
	"time"
)

const backupTimeLayout = "20060102T150405.000000000Z"

var backupPathPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$`)

// Run 分发备份计划，来源在执行时逐项校验，单项失败不阻止其他来源。
func (s *Manager) Run(ctx context.Context, id int64, content string, writeLog func(string)) error {
	if writeLog == nil {
		writeLog = func(string) {}
	}
	config, err := s.parseConfig(content)
	if err == nil {
		config, err = config.Validate()
	}
	if err != nil {
		writeLog("备份计划失败: " + err.Error())
		return err
	}
	switch config.Type {
	case "file", "directory":
		var options PathOptions
		options, err = s.parsePathOptions(config.Options)
		if err == nil {
			return s.RunPaths(ctx, id, options.Paths, config.Type == "directory", config, writeLog)
		}
	case "site":
		return s.runSites(ctx, id, config, writeLog)
	default:
		err = errors.New("备份类型不支持")
	}
	writeLog("备份计划失败: " + err.Error())
	return err
}

// validateBackupPath 限制类型及对象标识为单个目录名，本地和 OSS 使用相同隔离规则。
func (s *Manager) validateBackupPath(key, id string) error {
	if !backupPathPattern.MatchString(key) || !backupPathPattern.MatchString(id) {
		return errors.New("备份类型或对象标识不合法")
	}
	return nil
}

// newBackupName 使用 UTC 纳秒时间、计划 ID 和随机标识生成备份文件名。
func (s *Manager) newBackupName(planID int64, extension string) (string, error) {
	if planID <= 0 {
		return "", errors.New("备份计划 ID 不合法")
	}
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return "v2_" + time.Now().UTC().Format(backupTimeLayout) + "_" + strconv.FormatInt(planID, 10) + "_" + hex.EncodeToString(value[:]) + extension, nil
}

// backupDirectory 固定根目录及两层业务目录的句柄，拒绝符号链接和目录替换。
func (s *Manager) backupDirectory(key, id string, create bool) (*os.Root, error) {
	if err := s.validateBackupPath(key, id); err != nil {
		return nil, err
	}
	if create {
		if err := os.MkdirAll(s.root, 0700); err != nil {
			return nil, err
		}
	}
	resolved, err := filepath.EvalSymlinks(s.root)
	if err != nil {
		return nil, err
	}
	if filepath.Clean(resolved) != filepath.Clean(s.root) {
		return nil, errors.New("备份目录不能包含符号链接")
	}
	info, err := os.Lstat(s.root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("备份根路径必须是普通目录")
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return nil, err
	}
	current, err := root.Stat(".")
	if err != nil || !os.SameFile(info, current) {
		_ = root.Close()
		return nil, errors.New("备份根目录已变化")
	}
	defer func() {
		if root != nil {
			_ = root.Close()
		}
	}()
	for _, name := range []string{key, id} {
		if create {
			if err = root.Mkdir(name, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				return nil, err
			}
		}
		info, err := root.Lstat(name)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("备份路径必须是普通目录")
		}
		directory, err := root.OpenRoot(name)
		if err != nil {
			return nil, err
		}
		current, err := directory.Stat(".")
		if err != nil || !os.SameFile(info, current) {
			_ = directory.Close()
			return nil, errors.New("备份目录已变化")
		}
		_ = root.Close()
		root = directory
	}
	directory := root
	root = nil
	return directory, nil
}

// backupPrefix 按备份类型和对象隔离 OSS 文件，保留结尾分隔符避免误匹配。
func (s *Manager) backupPrefix(key, id string, config Config) string {
	prefix := strings.Trim(config.Prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return prefix + key + "/" + id + "/"
}

// Save 将调用方输出保存为完整文件后发布，成功后才清理超出的副本。
// write 必须同步写入并响应 ctx；本方法不启动后台任务。
func (s *Manager) Save(ctx context.Context, target Target, config Config, write func(context.Context, io.Writer) error) (result *File, err error) {
	if write == nil {
		return nil, errors.New("备份文件写入方法不能为空")
	}
	if target, err = s.validateTarget(target); err != nil {
		return nil, err
	}
	if config, err = config.Validate(); err != nil {
		return nil, err
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
	directory, err := s.backupDirectory(target.Key, target.ID, true)
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	name, err := s.newBackupName(target.PlanId, target.Extension)
	if err != nil {
		return nil, err
	}
	temporary := ".pending-" + name
	file, err := directory.OpenFile(temporary, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
		if removeErr := directory.Remove(temporary); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			if result != nil && err == nil {
				result.Warning = strings.TrimSpace(result.Warning + " 清理临时备份失败: " + removeErr.Error())
			} else {
				err = errors.Join(err, fmt.Errorf("清理临时备份失败: %w", removeErr))
			}
		}
	}()
	if err = write(ctx, file); err != nil {
		return nil, fmt.Errorf("生成备份失败: %w", err)
	}
	if err = file.Sync(); err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	result, err = s.parseBackupFile(name, info.Size(), config.Storage, target.Extension)
	if err != nil {
		return nil, err
	}
	if config.Storage == "oss" {
		// 先完成本地备份，再读取云端配置；实例删除不影响文件备份。
		var storageErr error
		if s.ossService == nil {
			storageErr = errors.New("对象存储服务未初始化")
		} else {
			storageErr = s.ossService.Check(config.OssId)
		}
		if storageErr == nil {
			key := s.backupPrefix(target.Key, target.ID, config) + name
			if _, err = file.Seek(0, 0); err != nil {
				return nil, err
			}
			_, err = s.ossService.Upload(ctx, config.OssId, key, serviceOss.UploadOptions{
				Body:          file,
				Size:          info.Size(),
				UploadOptions: oss.UploadOptions{PutOptions: oss.PutOptions{ContentType: target.ContentType}},
			})
			if err == nil {
				var uploadedFile *oss.File
				uploadedFile, err = s.ossService.StatFile(ctx, config.OssId, key)
				if err != nil {
					err = fmt.Errorf("上传完成，核对文件信息失败: %w", err)
				} else if uploadedFile == nil || uploadedFile.Size != info.Size() {
					err = errors.New("上传后的备份大小不匹配")
				}
			}
			if err != nil {
				// 上传期间删除实例时仍保留本地；其他上传错误正常返回。
				storageErr = s.ossService.Check(config.OssId)
				if storageErr == nil {
					return nil, err
				}
			}
		}
		if storageErr != nil {
			config.Storage = "local"
			config.LocalKeep = max(1, config.LocalKeep)
			result.Storage = "local"
			result.Warning = "云端存储不可用，已保留本地备份: " + storageErr.Error()
		}
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	// 云端不可用时至少保留一份；上传成功则遵循原有本地副本策略。
	if config.LocalKeep > 0 {
		if err = directory.Rename(temporary, name); err != nil {
			return nil, err
		}
	}
	local := config
	local.Storage = "local"
	localName := name
	if config.LocalKeep == 0 {
		localName = ""
	}
	cleanupErr := s.retainBackupFiles(ctx, target, local, localName)
	if cleanupErr != nil {
		cleanupErr = fmt.Errorf("清理本地副本失败: %w", cleanupErr)
	}
	if config.Storage == "oss" {
		if err = s.retainBackupFiles(ctx, target, config, name); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("清理云端副本失败: %w", err))
		}
	}
	if cleanupErr != nil {
		result.Warning = strings.TrimSpace(result.Warning + " 备份成功，" + cleanupErr.Error())
	}
	return result, nil
}

// retainBackupFiles 完整读取列表后按计划保留副本，本次发布的文件始终保留。
func (s *Manager) retainBackupFiles(ctx context.Context, target Target, config Config, published string) error {
	files, err := s.selectBackupFiles(ctx, target, config)
	if err != nil {
		return err
	}
	keep := config.LocalKeep
	if config.Storage == "oss" {
		keep = config.OssKeep
	}
	if keep > 0 {
		found := false
		for _, file := range files {
			if file.Name == published {
				found = true
				break
			}
		}
		if !found {
			return errors.New("新备份未出现在文件列表中，已保留旧副本")
		}
	}
	kept := 0
	if published != "" {
		kept = 1
	}
	for _, file := range files {
		if file.Name == published || file.PlanId != target.PlanId {
			continue
		}
		if kept < keep {
			kept++
			continue
		}
		if err = s.deleteBackupFile(ctx, target, config, file.Name); err != nil {
			return err
		}
	}
	return nil
}
