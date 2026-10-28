package oss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxMemoryPartSize int64 = 64 << 20 // 内存分片上限，超出后使用临时文件

// Upload 从流上传，size=-1 表示长度未知，可直接接收压缩或备份管道。
// 自动上传不保留失败分片，取消或错误时会独立超时清理；需要断点续传请使用分片 API。
// 对可能永久阻塞的 reader（例如 io.Pipe），调用方应在取消时关闭 reader 或停止生产者。
func (m *Manager) Upload(ctx context.Context, key string, body io.Reader, size int64, options UploadOptions) (_ *Object, err error) {
	if err = m.objectReady(ctx, key); err != nil {
		return nil, err
	}
	if body == nil || size < -1 || size > MaxObjectSize {
		return nil, invalid("上传需要 reader，长度必须为 -1 或 0..5 TiB")
	}
	partSize := options.PartSize
	if partSize == 0 {
		partSize = DefaultPartSize
	}
	if partSize < MinPartSize || partSize > MaxPartSize {
		return nil, invalid("PartSize 必须为 5 MiB..5 GiB")
	}
	if size > 0 && (size+partSize-1)/partSize > MaxParts {
		partSize = (size + MaxParts - 1) / MaxParts
	}
	reader := &contextReader{ctx: ctx, reader: body}
	buffer := uploadPartBuffer{tempDir: options.TempDir}
	firstSize := partSize
	if size >= 0 && size < firstSize {
		firstSize = size
	}
	first, n, readErr := buffer.read(reader, firstSize)
	if first != nil {
		defer func() {
			err = errors.Join(err, first.Close())
		}()
	}
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return nil, readErr
	}
	if size >= 0 && n < firstSize {
		return nil, io.ErrUnexpectedEOF
	}
	if (size >= 0 && size <= partSize) || (size == -1 && errors.Is(readErr, io.EOF)) {
		object, putErr := m.PutObject(ctx, key, first, n, options.PutOptions)
		if putErr == nil && options.Progress != nil {
			options.Progress(n, size)
		}
		return object, putErr
	}
	upload, err := m.InitiateMultipartUpload(ctx, key, options.PutOptions)
	if err != nil {
		return nil, err
	}
	completed := false
	defer func() {
		if !completed {
			err = errors.Join(err, m.abortUpload(ctx, key, upload.UploadID))
		}
	}()
	partCount := 0
	if size > 0 {
		partCount = int((size + partSize - 1) / partSize)
	}
	parts := make([]Part, 0, partCount)
	var uploaded int64
	partBody := first
	for {
		if n > 0 {
			if len(parts) >= MaxParts || uploaded+n > MaxObjectSize {
				return nil, errors.Join(invalid("上传超过最大分片数或对象大小，请增大 PartSize"), partBody.Close())
			}
			part, partErr := m.UploadPart(ctx, key, upload.UploadID, len(parts)+1, partBody, n)
			closeErr := partBody.Close()
			if partErr != nil || closeErr != nil {
				return nil, errors.Join(partErr, closeErr)
			}
			parts = append(parts, *part)
			uploaded += n
			if options.Progress != nil {
				options.Progress(uploaded, size)
			}
		}
		if (size >= 0 && uploaded == size) || errors.Is(readErr, io.EOF) {
			break
		}
		nextSize := partSize
		if size >= 0 && size-uploaded < nextSize {
			nextSize = size - uploaded
		}
		partBody, n, readErr = buffer.read(reader, nextSize)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil, errors.Join(readErr, partBody.Close())
		}
		if size >= 0 && n < nextSize {
			return nil, errors.Join(io.ErrUnexpectedEOF, partBody.Close())
		}
		if n == 0 {
			if closeErr := partBody.Close(); closeErr != nil {
				return nil, closeErr
			}
		}
	}
	object, err := m.CompleteMultipartUpload(ctx, key, upload.UploadID, parts)
	if err != nil {
		return nil, err
	}
	completed = true
	object.Size = uploaded
	return object, nil
}

// UploadFile 上传本地普通文件。文件在上传期间应保持不变。
func (m *Manager) UploadFile(ctx context.Context, key, filename string, options UploadOptions) (*Object, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("读取上传文件信息失败: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, invalid("只支持上传普通文件")
	}
	return m.Upload(ctx, key, file, info.Size(), options)
}

// DownloadFile 将完整对象下载到同目录的临时文件，成功后替换目标，失败保留原文件。
func (m *Manager) DownloadFile(ctx context.Context, key, filename string) (err error) {
	download, err := m.GetObject(ctx, key, GetOptions{})
	if err != nil {
		return err
	}
	defer download.Body.Close()
	file, err := os.CreateTemp(filepath.Dir(filename), ".oss-download-*")
	if err != nil {
		return fmt.Errorf("创建下载临时文件失败: %w", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	n, err := io.Copy(file, &contextReader{ctx: ctx, reader: download.Body})
	if err != nil {
		return fmt.Errorf("下载对象失败: %w", err)
	}
	if download.Object.Size >= 0 && n != download.Object.Size {
		return io.ErrUnexpectedEOF
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filename)
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

// Read 在每次读取前检查取消状态；底层阻塞读取由调用方负责中断。
func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

// uploadPartReader 支持 SDK 重试时回退读取，并负责当前分片的临时文件。
type uploadPartReader struct {
	io.ReadSeeker
	file   *os.File
	closed bool
}

// Close 释放当前分片及临时文件，允许对空值和已关闭的分片重复调用。
func (r *uploadPartReader) Close() error {
	if r == nil || r.closed {
		return nil
	}
	r.closed = true
	r.ReadSeeker = nil
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	return errors.Join(err, os.Remove(r.file.Name()))
}

// uploadPartBuffer 在单次顺序上传中复用内存，前一分片关闭后才读取下一分片。
type uploadPartBuffer struct {
	memory  []byte
	tempDir string
}

// read 复用当前内存分片，大分片转存磁盘，避免每上传一片就重新分配大块内存。
func (b *uploadPartBuffer) read(reader io.Reader, size int64) (*uploadPartReader, int64, error) {
	if size <= maxMemoryPartSize {
		if int64(cap(b.memory)) < size {
			b.memory = make([]byte, int(size))
		}
		buffer := b.memory[:int(size)]
		n, err := io.ReadFull(reader, buffer)
		if errors.Is(err, io.ErrUnexpectedEOF) {
			err = io.EOF
		}
		return &uploadPartReader{ReadSeeker: bytes.NewReader(buffer[:n])}, int64(n), err
	}
	file, err := os.CreateTemp(b.tempDir, ".oss-upload-*")
	if err != nil {
		return nil, 0, err
	}
	part := &uploadPartReader{ReadSeeker: file, file: file}
	n, err := io.CopyN(file, reader, size)
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		err = errors.Join(err, seekErr)
	}
	return part, n, err
}
