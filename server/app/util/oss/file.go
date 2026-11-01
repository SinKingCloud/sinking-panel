package oss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ListFiles 按前缀、目录分隔符和游标分页查询文件。
func (m *Manager) ListFiles(ctx context.Context, o ListOptions) (*FilePage, error) {
	if err := m.ready(ctx, true); err != nil {
		return nil, err
	}
	var err error
	if o.Limit, err = m.pageLimit(o.Limit); err != nil {
		return nil, err
	}
	page, err := m.driver.ListFiles(ctx, o)
	if err != nil {
		return nil, err
	}
	if page.HasMore && (page.NextToken == "" || page.NextToken == o.Token) {
		return nil, m.providerError("ListFiles", o.Prefix, io.ErrNoProgress)
	}
	return page, nil
}

// StatFile 读取文件大小、时间和元数据，不下载正文。
func (m *Manager) StatFile(ctx context.Context, key string) (*File, error) {
	if err := m.fileReady(ctx, key); err != nil {
		return nil, err
	}
	return m.driver.StatFile(ctx, key)
}

// Exists 判断文件是否存在，仅将 404 视为不存在。
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.StatFile(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

// GetFile 打开下载流，Progress 按读取字节通知；调用方必须关闭返回的 Body。
func (m *Manager) GetFile(ctx context.Context, key string, o GetOptions) (*Download, error) {
	if err := m.fileReady(ctx, key); err != nil {
		return nil, err
	}
	if o.Range != "" && !regexp.MustCompile(`^bytes=(\d+-\d*|-\d+)$`).MatchString(o.Range) {
		return nil, m.invalid("Range 必须是单个 bytes 范围")
	}
	callback := o.Progress
	o.Progress = nil
	download, err := m.driver.GetFile(ctx, key, o)
	if err != nil {
		return nil, err
	}
	if callback != nil {
		download.Body = &progressReadCloser{
			ctx:      ctx,
			body:     download.Body,
			progress: m.newTransferProgress(ctx, download.File.Size, callback),
		}
	}
	return download, nil
}

// PutFile 上传已知长度的小文件；大文件和未知长度流使用 Upload。
func (m *Manager) PutFile(ctx context.Context, key string, body io.Reader, size int64, o PutOptions) (*File, error) {
	if err := m.fileReady(ctx, key); err != nil {
		return nil, err
	}
	if body == nil || size < 0 || size > MaxPartSize {
		return nil, m.invalid("简单上传需要 reader 和 0..5 GiB 长度，大文件请用 Upload")
	}
	return m.driver.PutFile(ctx, key, body, size, o)
}

// DeleteFile 删除指定文件。
func (m *Manager) DeleteFile(ctx context.Context, key string) error {
	if err := m.fileReady(ctx, key); err != nil {
		return err
	}
	return m.driver.DeleteFile(ctx, key)
}

// CopyFile 在当前存储桶内复制文件，超过 1 GiB 时自动使用服务端分片复制。
func (m *Manager) CopyFile(ctx context.Context, source, destination string, o CopyOptions) (*File, error) {
	if err := m.fileReady(ctx, source); err != nil {
		return nil, err
	}
	if err := m.validKey(destination); err != nil {
		return nil, err
	}
	file, err := m.StatFile(ctx, source)
	if err != nil {
		return nil, err
	}
	// 采用 1 GiB 作为服务端单次复制阈值，覆盖五家厂商的共同能力。
	if file.Size <= 1<<30 {
		result, err := m.driver.CopyFile(ctx, source, destination, o)
		if err != nil {
			return nil, err
		}
		result.Size = file.Size
		return result, nil
	}
	return m.copyMultipart(ctx, source, destination, file, o)
}

// MoveFile 先复制成功再删除源；不是原子操作，删除失败时两份文件都保留。
// 源和目标位于当前存储桶，调用期间源文件必须保持不变。
func (m *Manager) MoveFile(ctx context.Context, source, destination string) (*File, error) {
	if source == destination {
		return nil, m.invalid("源和目标不能相同")
	}
	file, err := m.CopyFile(ctx, source, destination, CopyOptions{})
	if err != nil {
		return nil, err
	}
	if err = m.DeleteFile(ctx, source); err != nil {
		return file, err
	}
	return file, nil
}

// RenameFile 在原目录内重命名文件，name 只包含新文件名，不改变目录前缀。
// 先检查同名目标，再复用云端复制和删除；支持大文件，但不是原子操作。
func (m *Manager) RenameFile(ctx context.Context, key, name string) (*File, error) {
	if err := m.fileReady(ctx, key); err != nil {
		return nil, err
	}
	if strings.HasSuffix(key, "/") {
		return nil, m.invalid("重命名只支持文件，不支持目录")
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
		return nil, m.invalid("请输入新文件名，不能包含目录分隔符")
	}
	destination := key[:strings.LastIndex(key, "/")+1] + name
	if destination == key {
		return nil, m.invalid("新名称不能与原名称相同")
	}
	exists, err := m.Exists(ctx, destination)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: 目标文件已存在", ErrConflict)
	}
	return m.MoveFile(ctx, key, destination)
}

// SetMetadata 整体替换用户元数据及可写内容头，未传入的字段不会保留。
func (m *Manager) SetMetadata(ctx context.Context, key string, o PutOptions) (*File, error) {
	return m.CopyFile(ctx, key, key, CopyOptions{ReplaceMetadata: true, PutOptions: o})
}

// CreateDirectory 创建以斜杠结尾的空文件，作为虚拟目录标记。
func (m *Manager) CreateDirectory(ctx context.Context, prefix string) (*File, error) {
	if prefix == "" {
		return nil, m.invalid("目录不能为空")
	}
	return m.PutFile(ctx, strings.TrimSuffix(prefix, "/")+"/", strings.NewReader(""), 0, PutOptions{ContentType: "application/x-directory"})
}

// DeleteFiles 返回逐项结果；任何失败也返回聚合错误，便于备份清理检查部分成功。
func (m *Manager) DeleteFiles(ctx context.Context, keys []string) ([]DeleteResult, error) {
	if err := m.ready(ctx, true); err != nil {
		return nil, err
	}
	for _, key := range keys {
		if err := m.validKey(key); err != nil {
			return nil, err
		}
	}
	results := make([]DeleteResult, 0, len(keys))
	var failures []error
	for _, key := range keys {
		err := m.DeleteFile(ctx, key)
		results = append(results, DeleteResult{Key: key, Err: err})
		if err != nil {
			failures = append(failures, err)
		}
	}
	return results, errors.Join(failures...)
}
