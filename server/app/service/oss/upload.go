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

// Upload 统一处理文件流、分片会话和服务器后台上传，直接覆盖同名文件，返回对应操作结果或任务 ID。
// 默认同步上传并自动分片；server 排队后重新检查文件，使用后台任务的上下文和进度。
func (s *service) Upload(ctx context.Context, id int64, key string, options UploadOptions) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch options.Action {
	case "", "file", "init", "part", "complete", "abort", "parts", "uploads":
		client, err := s.client(id)
		if err != nil {
			return nil, err
		}
		switch options.Action {
		case "init":
			return client.InitiateMultipartUpload(ctx, key, options.PutOptions)
		case "part":
			return client.UploadPart(ctx, key, options.UploadId, options.Number, options.Body, options.Size)
		case "complete":
			return client.CompleteMultipartUpload(ctx, key, options.UploadId, options.Parts)
		case "abort":
			return nil, client.AbortMultipartUpload(ctx, key, options.UploadId)
		case "parts":
			return client.ListParts(ctx, key, options.UploadId, utilOss.PartsOptions{Marker: options.Marker, Limit: options.Limit})
		case "uploads":
			return client.ListMultipartUploads(ctx, utilOss.UploadsOptions{
				Prefix: options.Prefix, KeyMarker: options.KeyMarker, UploadIDMarker: options.UploadIdMarker, Limit: options.Limit,
			})
		default:
			return client.Upload(ctx, key, options.Body, options.Size, options.UploadOptions)
		}
	case "server":
	default:
		return nil, errors.New("上传操作类型不合法")
	}
	path := options.Path
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
		return nil, err
	}
	options.Metadata = maps.Clone(options.Metadata)
	return s.transferTask(id, "upload", "上传", key, path, func(ctx context.Context, progress utilOss.ProgressFunc) (int64, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := check(); err != nil {
			return 0, err
		}
		client, err := s.client(id)
		if err != nil {
			return 0, err
		}
		upload := options.UploadOptions
		upload.Progress = progress
		file, err := client.UploadFile(ctx, key, path, upload)
		if err != nil {
			return 0, err
		}
		return file.Size, nil
	})
}
