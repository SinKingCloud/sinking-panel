package backup

import (
	"context"
	"errors"
	"io"
	"os"
	"regexp"
	serviceOss "server/app/service/oss"
	"server/app/util/page"
	"sort"
	"strconv"
	"strings"
	"time"
)

// parseBackupFile 只接受当前命名规则和指定扩展名，不将未知文件纳入管理。
func (s *Manager) parseBackupFile(name string, size int64, storage, extension string) (*File, error) {
	if extension == "" || !strings.HasSuffix(name, extension) {
		return nil, errors.New("备份文件扩展名不合法")
	}
	match := regexp.MustCompile(`^v2_([0-9]{8}T[0-9]{6}\.[0-9]{9}Z)_([1-9][0-9]*)_[0-9a-f]{32}$`).FindStringSubmatch(strings.TrimSuffix(name, extension))
	if len(match) != 3 {
		return nil, errors.New("备份文件名不合法")
	}
	created, err := time.Parse(backupTimeLayout, match[1])
	if err != nil {
		return nil, errors.New("备份文件时间不合法")
	}
	planID, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		return nil, errors.New("备份计划 ID 不合法")
	}
	return &File{
		Name:    name,
		Size:    size,
		Time:    created.UnixMilli(),
		Storage: storage,
		PlanId:  planID,
	}, nil
}

// selectBackupFiles 读取完整列表并按 UTC 时间倒序排列，供分页和副本清理共用。
func (s *Manager) selectBackupFiles(ctx context.Context, target Target, config Config) ([]*File, error) {
	if err := s.validateBackupPath(target.Key, target.ID); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	files := make([]*File, 0)
	if config.Storage == "oss" {
		prefix := s.backupPrefix(target.Key, target.ID, config)
		query := &page.Query{PageSize: 1000}
		for {
			data, err := s.ossService.Files(ctx, config.OssId, prefix, "", query)
			if err != nil {
				return nil, err
			}
			if data == nil {
				return nil, errors.New("OSS 备份列表为空响应")
			}
			for _, item := range data.List {
				if item == nil || !strings.HasPrefix(item.Key, prefix) {
					continue
				}
				file, err := s.parseBackupFile(strings.TrimPrefix(item.Key, prefix), item.Size, config.Storage, target.Extension)
				if err == nil {
					files = append(files, file)
				}
			}
			if data.HasMore == nil || !*data.HasMore {
				break
			}
			if data.NextCursorId == nil || *data.NextCursorId == "" || *data.NextCursorId == query.CursorId {
				return nil, errors.New("OSS 备份列表分页游标不合法")
			}
			query.CursorId = *data.NextCursorId
		}
	} else if config.Storage == "local" {
		directory, err := s.backupDirectory(target.Key, target.ID, false)
		if errors.Is(err, os.ErrNotExist) {
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		defer directory.Close()
		entries, err := directory.Open(".")
		if err != nil {
			return nil, err
		}
		defer entries.Close()
		for {
			batch, err := entries.ReadDir(256)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}
			for _, entry := range batch {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				file, parseErr := s.parseBackupFile(entry.Name(), 0, config.Storage, target.Extension)
				if parseErr != nil || !entry.Type().IsRegular() {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					return nil, err
				}
				if info.Mode().IsRegular() {
					file.Size = info.Size()
					files = append(files, file)
				}
			}
			if errors.Is(err, io.EOF) {
				break
			}
		}
	} else {
		return nil, errors.New("备份存储类型不支持")
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name > files[j].Name })
	return files, nil
}

// deleteBackupFile 删除指定备份，拒绝路径、目录和符号链接。
func (s *Manager) deleteBackupFile(ctx context.Context, target Target, config Config, name string) error {
	if err := s.validateBackupPath(target.Key, target.ID); err != nil {
		return err
	}
	if _, err := s.parseBackupFile(name, 0, config.Storage, target.Extension); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if config.Storage == "oss" {
		_, err := s.ossService.Delete(ctx, config.OssId, []string{s.backupPrefix(target.Key, target.ID, config) + name})
		return err
	}
	if config.Storage != "local" {
		return errors.New("备份存储类型不支持")
	}
	directory, err := s.backupDirectory(target.Key, target.ID, false)
	if err != nil {
		return err
	}
	defer directory.Close()
	info, err := directory.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("备份路径不是普通文件")
	}
	return directory.Remove(name)
}

// openBackupFile 打开恢复所需的数据流，调用者负责关闭，本地文件不得被替换为链接。
func (s *Manager) openBackupFile(ctx context.Context, target Target, config Config, name string) (io.ReadCloser, error) {
	if err := s.validateBackupPath(target.Key, target.ID); err != nil {
		return nil, err
	}
	if _, err := s.parseBackupFile(name, 0, config.Storage, target.Extension); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if config.Storage == "oss" {
		download, err := s.ossService.Download(ctx, config.OssId, s.backupPrefix(target.Key, target.ID, config)+name, serviceOss.DownloadOptions{})
		if err != nil {
			return nil, err
		}
		if download == nil || download.Download == nil || download.Body == nil {
			return nil, errors.New("OSS 备份下载响应无效")
		}
		return download.Body, nil
	}
	if config.Storage != "local" {
		return nil, errors.New("备份存储类型不支持")
	}
	directory, err := s.backupDirectory(target.Key, target.ID, false)
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	info, err := directory.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("备份路径不是普通文件")
	}
	file, err := directory.Open(name)
	if err != nil {
		return nil, err
	}
	current, err := file.Stat()
	if err != nil || !os.SameFile(info, current) || !current.Mode().IsRegular() {
		_ = file.Close()
		return nil, errors.New("备份文件已变化")
	}
	return file, nil
}
