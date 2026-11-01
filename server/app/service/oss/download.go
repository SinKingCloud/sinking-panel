package oss

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	utilOss "server/app/util/oss"
	"strings"
	"unicode/utf8"
)

// Download 将 OSS 文件下载加入后台任务，写完临时文件后才发布到目标路径。
func (s *service) Download(id int64, key, path string, overwrite bool) (string, error) {
	check := func() error {
		if key == "" || len(key) > 1024 || !utf8.ValidString(key) || strings.ContainsAny(key, "\x00\r\n") || strings.HasSuffix(key, "/") {
			return errors.New("请选择有效的 OSS 文件名称")
		}
		if !filepath.IsAbs(path) || len(path) > 4096 || strings.ContainsRune(path, 0) {
			return errors.New("请填写服务器上的完整文件路径")
		}
		parent, err := os.Stat(filepath.Dir(path))
		if err != nil {
			return err
		}
		if !parent.IsDir() {
			return errors.New("目标文件的上级路径必须是目录")
		}
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("下载目标必须是普通文件，不能是目录或符号链接")
		}
		if !overwrite {
			return errors.New("目标文件已存在，请确认是否覆盖")
		}
		return nil
	}
	if err := check(); err != nil {
		return "", err
	}
	return s.transferTask(id, "download", "下载", key, path, overwrite, func(ctx context.Context, progress utilOss.ProgressFunc) (int64, error) {
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
		file, err := os.CreateTemp(filepath.Dir(path), ".oss-transfer-*")
		if err != nil {
			return 0, err
		}
		filename := file.Name()
		defer os.Remove(filename)
		if err = file.Close(); err != nil {
			return 0, err
		}
		if err = client.DownloadFile(ctx, key, filename, utilOss.DownloadOptions{Progress: progress}); err != nil {
			return 0, err
		}
		info, err := os.Stat(filename)
		if err != nil {
			return 0, err
		}
		if err = ctx.Err(); err != nil {
			return 0, err
		}
		if err = check(); err != nil {
			return 0, err
		}
		// 同目录重命名发布；禁止覆盖时用硬链接阻止并发写入同一目标。
		if overwrite {
			err = os.Rename(filename, path)
		} else {
			err = os.Link(filename, path)
		}
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	})
}
