package oss

import (
	"context"
	"errors"
	"io"
	"strings"
)

// ListObjects 按前缀、目录分隔符和游标分页查询文件。
func (m *Manager) ListObjects(ctx context.Context, o ListOptions) (*ObjectPage, error) {
	if err := m.ready(ctx, true); err != nil {
		return nil, err
	}
	var err error
	if o.Limit, err = pageLimit(o.Limit); err != nil {
		return nil, err
	}
	page, err := m.driver.ListObjects(ctx, o)
	if err != nil {
		return nil, err
	}
	if page.HasMore && (page.NextToken == "" || page.NextToken == o.Token) {
		return nil, providerError(m.config.Provider, "ListObjects", o.Prefix, io.ErrNoProgress)
	}
	return page, nil
}

// StatObject 读取文件大小、时间和元数据，不下载正文。
func (m *Manager) StatObject(ctx context.Context, key string) (*Object, error) {
	if err := m.objectReady(ctx, key); err != nil {
		return nil, err
	}
	return m.driver.StatObject(ctx, key)
}

// Exists 判断对象是否存在，仅将 404 视为不存在。
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.StatObject(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

// GetObject 打开下载流，调用方必须关闭返回的 Body。
func (m *Manager) GetObject(ctx context.Context, key string, o GetOptions) (*Download, error) {
	if err := m.objectReady(ctx, key); err != nil {
		return nil, err
	}
	if o.Range != "" && !objectRangePattern.MatchString(o.Range) {
		return nil, invalid("Range 必须是单个 bytes 范围")
	}
	return m.driver.GetObject(ctx, key, o)
}

// PutObject 上传已知长度的小文件；大文件和未知长度流使用 Upload。
func (m *Manager) PutObject(ctx context.Context, key string, body io.Reader, size int64, o PutOptions) (*Object, error) {
	if err := m.objectReady(ctx, key); err != nil {
		return nil, err
	}
	if body == nil || size < 0 || size > MaxPartSize {
		return nil, invalid("简单上传需要 reader 和 0..5 GiB 长度，大文件请用 Upload")
	}
	return m.driver.PutObject(ctx, key, body, size, o)
}

// DeleteObject 删除指定对象。
func (m *Manager) DeleteObject(ctx context.Context, key string) error {
	if err := m.objectReady(ctx, key); err != nil {
		return err
	}
	return m.driver.DeleteObject(ctx, key)
}

// CopyObject 复制对象，超过 1 GiB 时自动使用服务端分片复制。
func (m *Manager) CopyObject(ctx context.Context, source, destination string, o CopyOptions) (*Object, error) {
	if err := m.objectReady(ctx, source); err != nil {
		return nil, err
	}
	if err := validKey(destination); err != nil {
		return nil, err
	}
	sourceManager := m
	if o.SourceBucket != "" && o.SourceBucket != m.config.Bucket {
		if !bucketNamePattern.MatchString(o.SourceBucket) {
			return nil, invalid("源桶名称不合法")
		}
		config := m.config
		config.Bucket = o.SourceBucket
		var err error
		// 沿用已校验的连接配置，复用 HTTP 连接池。
		sourceManager, err = newManager(config)
		if err != nil {
			return nil, err
		}
	}
	object, err := sourceManager.StatObject(ctx, source)
	if err != nil {
		return nil, err
	}
	// 采用 1 GiB 作为服务端单次复制阈值，覆盖五家厂商的共同能力。
	if object.Size <= 1<<30 {
		result, err := m.driver.CopyObject(ctx, source, destination, o)
		if err != nil {
			return nil, err
		}
		result.Size = object.Size
		return result, nil
	}
	return m.copyMultipart(ctx, source, destination, object, o)
}

// MoveObject 先复制成功再删除源；不是原子操作，删除失败时两份对象都保留。
// 调用期间源对象必须保持不变；跨桶移动使用独立的复制/删除操作。
func (m *Manager) MoveObject(ctx context.Context, source, destination string) (*Object, error) {
	if source == destination {
		return nil, invalid("源和目标不能相同")
	}
	object, err := m.CopyObject(ctx, source, destination, CopyOptions{})
	if err != nil {
		return nil, err
	}
	if err = m.DeleteObject(ctx, source); err != nil {
		return object, err
	}
	return object, nil
}

// SetMetadata 整体替换用户元数据及可写内容头，未传入的字段不会保留。
func (m *Manager) SetMetadata(ctx context.Context, key string, o PutOptions) (*Object, error) {
	return m.CopyObject(ctx, key, key, CopyOptions{ReplaceMetadata: true, PutOptions: o})
}

// CreateDirectory 创建以斜杠结尾的空对象，作为虚拟目录标记。
func (m *Manager) CreateDirectory(ctx context.Context, prefix string) (*Object, error) {
	if prefix == "" {
		return nil, invalid("目录不能为空")
	}
	return m.PutObject(ctx, strings.TrimSuffix(prefix, "/")+"/", strings.NewReader(""), 0, PutOptions{ContentType: "application/x-directory"})
}

// DeleteObjects 返回逐项结果；任何失败也返回聚合错误，便于备份清理检查部分成功。
func (m *Manager) DeleteObjects(ctx context.Context, keys []string) ([]DeleteResult, error) {
	if err := m.ready(ctx, true); err != nil {
		return nil, err
	}
	for _, key := range keys {
		if err := validKey(key); err != nil {
			return nil, err
		}
	}
	results := make([]DeleteResult, 0, len(keys))
	var failures []error
	for _, key := range keys {
		err := m.DeleteObject(ctx, key)
		results = append(results, DeleteResult{Key: key, Err: err})
		if err != nil {
			failures = append(failures, err)
		}
	}
	return results, errors.Join(failures...)
}
