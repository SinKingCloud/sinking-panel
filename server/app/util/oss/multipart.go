package oss

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"
)

// UploadPartCopy 在云端复制源对象的指定字节范围，不经本地下载。
func (m *Manager) UploadPartCopy(ctx context.Context, key, id string, number int, source string, options CopyPartOptions) (*Part, error) {
	if err := m.multipartReady(ctx, key, id); err != nil {
		return nil, err
	}
	if err := validKey(source); err != nil {
		return nil, err
	}
	if number < 1 || number > MaxParts || options.Offset < 0 || options.Size <= 0 || options.Size > MaxPartSize || options.Offset > MaxObjectSize-options.Size {
		return nil, invalid("复制分片编号或字节范围不合法")
	}
	return m.driver.UploadPartCopy(ctx, key, id, number, source, options)
}

// copyMultipart 用 ETag 固定源对象，成功合并前不触碰已有目标对象。
func (m *Manager) copyMultipart(ctx context.Context, source, destination string, object *Object, options CopyOptions) (_ *Object, err error) {
	if object.Size > MaxObjectSize {
		return nil, invalid("对象超过 5 TiB")
	}
	if object.ETag == "" {
		return nil, invalid("分片复制需要源对象 ETag")
	}
	putOptions := options.PutOptions
	if !options.ReplaceMetadata {
		putOptions = PutOptions{
			ContentType:        object.ContentType,
			ContentDisposition: object.ContentDisposition,
			ContentEncoding:    object.ContentEncoding,
			CacheControl:       object.CacheControl,
			Metadata:           object.Metadata,
			StorageClass:       object.StorageClass,
		}
	}
	upload, err := m.InitiateMultipartUpload(ctx, destination, putOptions)
	if err != nil {
		return nil, err
	}
	completed := false
	defer func() {
		if completed {
			return
		}
		err = errors.Join(err, m.abortUpload(ctx, destination, upload.UploadID))
	}()
	partSize := max(int64(128<<20), (object.Size+MaxParts-1)/MaxParts)
	parts := make([]Part, 0, (object.Size+partSize-1)/partSize)
	for offset := int64(0); offset < object.Size; offset += partSize {
		size := min(partSize, object.Size-offset)
		part, copyErr := m.UploadPartCopy(ctx, destination, upload.UploadID, len(parts)+1, source, CopyPartOptions{
			SourceBucket: options.SourceBucket,
			Offset:       offset,
			Size:         size,
			IfMatch:      object.ETag,
		})
		if copyErr != nil {
			return nil, copyErr
		}
		parts = append(parts, *part)
	}
	result, err := m.CompleteMultipartUpload(ctx, destination, upload.UploadID, parts)
	if err != nil {
		return nil, err
	}
	completed = true
	result.Size = object.Size
	return result, nil
}

// InitiateMultipartUpload 初始化分片会话，调用方应保存 Key 和 UploadID 以便恢复。
func (m *Manager) InitiateMultipartUpload(ctx context.Context, key string, o PutOptions) (*MultipartUpload, error) {
	if err := m.objectReady(ctx, key); err != nil {
		return nil, err
	}
	return m.driver.InitiateMultipartUpload(ctx, key, o)
}

// multipartReady 检查分片操作的上下文、对象和会话标识。
func (m *Manager) multipartReady(ctx context.Context, key, id string) error {
	if err := m.objectReady(ctx, key); err != nil {
		return err
	}
	if id == "" {
		return invalid("UploadID 不能为空")
	}
	return nil
}

// UploadPart 上传一个分片，同编号重复上传会覆盖该分片。
func (m *Manager) UploadPart(ctx context.Context, key, id string, number int, body io.Reader, size int64) (*Part, error) {
	if err := m.multipartReady(ctx, key, id); err != nil {
		return nil, err
	}
	if number < 1 || number > MaxParts || body == nil || size <= 0 || size > MaxPartSize {
		return nil, invalid("分片编号必须为 1..10000，长度必须为 1..5 GiB")
	}
	return m.driver.UploadPart(ctx, key, id, number, body, size)
}

// CompleteMultipartUpload 按编号合并分片，要求提供各片 ETag 和准确长度。
func (m *Manager) CompleteMultipartUpload(ctx context.Context, key, id string, parts []Part) (*Object, error) {
	if err := m.multipartReady(ctx, key, id); err != nil {
		return nil, err
	}
	if len(parts) == 0 || len(parts) > MaxParts {
		return nil, invalid("分片数量必须为 1..10000")
	}
	// 自动上传通常已按编号排列，直接校验；乱序输入才复制排序，避免修改调用方切片。
	ordered := parts
	if !slices.IsSortedFunc(parts, compareParts) {
		ordered = slices.Clone(parts)
		slices.SortFunc(ordered, compareParts)
	}
	var total int64
	for i, p := range ordered {
		if p.Number != i+1 || p.ETag == "" || p.Size <= 0 || p.Size > MaxPartSize || (i < len(ordered)-1 && p.Size < MinPartSize) {
			return nil, invalid("分片需从 1 连续编号并包含 ETag、长度；非末片至少 5 MiB")
		}
		total += p.Size
	}
	if total > MaxObjectSize {
		return nil, invalid("对象超过 5 TiB")
	}
	return m.driver.CompleteMultipartUpload(ctx, key, id, ordered)
}

// AbortMultipartUpload 中止上传会话并释放尚未合并的分片。
func (m *Manager) AbortMultipartUpload(ctx context.Context, key, id string) error {
	if err := m.multipartReady(ctx, key, id); err != nil {
		return err
	}
	return m.driver.AbortMultipartUpload(ctx, key, id)
}

// abortUpload 使用独立超时清理失败会话，调用方取消后仍可回收服务端分片。
func (m *Manager) abortUpload(ctx context.Context, key, id string) error {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	if err := m.AbortMultipartUpload(cleanup, key, id); err != nil {
		return fmt.Errorf("清理上传会话 %s 失败: %w", id, err)
	}
	return nil
}

// compareParts 按分片编号排序，供顺序检查和合并使用。
func compareParts(left, right Part) int {
	return cmp.Compare(left.Number, right.Number)
}

// ListParts 按分片编号游标查询已上传分片。
func (m *Manager) ListParts(ctx context.Context, key, id string, o PartsOptions) (*PartsPage, error) {
	if err := m.multipartReady(ctx, key, id); err != nil {
		return nil, err
	}
	var err error
	if o.Limit, err = pageLimit(o.Limit); err != nil {
		return nil, err
	}
	if o.Marker < 0 || o.Marker > MaxParts {
		return nil, invalid("分片 Marker 不合法")
	}
	page, err := m.driver.ListParts(ctx, key, id, o)
	if err != nil {
		return nil, err
	}
	if page.HasMore && page.NextMarker <= o.Marker {
		return nil, providerError(m.config.Provider, "ListParts", key, io.ErrNoProgress)
	}
	return page, nil
}

// ListMultipartUploads 按双游标查询未完成的分片会话。
func (m *Manager) ListMultipartUploads(ctx context.Context, o UploadsOptions) (*UploadsPage, error) {
	if err := m.ready(ctx, true); err != nil {
		return nil, err
	}
	var err error
	if o.Limit, err = pageLimit(o.Limit); err != nil {
		return nil, err
	}
	page, err := m.driver.ListMultipartUploads(ctx, o)
	if err != nil {
		return nil, err
	}
	if page.HasMore && (page.NextKeyMarker == "" || (page.NextKeyMarker == o.KeyMarker && page.NextUploadIDMarker == o.UploadIDMarker)) {
		return nil, providerError(m.config.Provider, "ListMultipartUploads", o.Prefix, io.ErrNoProgress)
	}
	return page, nil
}
