package oss

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	utilOss "server/app/util/oss"
	"strings"
	"unicode/utf8"
)

// Upload 将服务器文件上传加入后台任务，排队后重新检查文件并保留上传进度。
func (s *service) Upload(id int64, key, path string, overwrite bool, options utilOss.PutOptions) (string, error) {
	check := func() error {
		if key == "" || len(key) > 1024 || !utf8.ValidString(key) || strings.ContainsAny(key, "\x00\r\n") || strings.HasSuffix(key, "/") {
			return errors.New("请选择有效的 OSS 文件名称")
		}
		if !filepath.IsAbs(path) || len(path) > 4096 || strings.ContainsRune(path, 0) {
			return errors.New("请填写服务器上的完整文件路径")
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("上传路径必须是普通文件")
		}
		if info.Size() > utilOss.MaxFileSize {
			return errors.New("上传文件大小不能超过 5 TiB")
		}
		return nil
	}
	if err := check(); err != nil {
		return "", err
	}
	options.Metadata = maps.Clone(options.Metadata)
	return s.transferTask(id, "upload", "上传", key, path, overwrite, func(ctx context.Context, progress utilOss.ProgressFunc) (int64, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := check(); err != nil {
			return 0, err
		}
		client, err := s.Client(id)
		if err != nil {
			return 0, err
		}
		if !overwrite {
			if _, err = client.StatFile(ctx, key); err == nil {
				return 0, errors.New("OSS 文件已存在，请确认是否覆盖")
			} else if !errors.Is(err, utilOss.ErrNotFound) {
				return 0, err
			}
		}
		file, err := client.UploadFile(ctx, key, path, utilOss.UploadOptions{
			PutOptions: options,
			Progress:   progress,
		})
		if err != nil {
			return 0, err
		}
		return file.Size, nil
	})
}
