package oss

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"
)

const progressInterval = 100 * time.Millisecond

// transferProgress 串行通知累计字节数，限制频率并将完成通知留给操作成功后发送。
type transferProgress struct {
	mu          sync.Mutex
	ctx         context.Context
	callback    ProgressFunc
	total       int64
	transferred int64
	reported    int64
	last        time.Time
	finished    bool
}

// newTransferProgress 仅在需要回调时创建状态；空对象在成功前不通知完成。
func (m *Manager) newTransferProgress(ctx context.Context, total int64, callback ProgressFunc) *transferProgress {
	if callback == nil {
		return nil
	}
	p := &transferProgress{
		ctx:      ctx,
		total:    total,
		callback: callback,
		last:     time.Now(),
	}
	if total != 0 && ctx.Err() == nil {
		callback(0, total)
	}
	return p
}

// update 按最高已传输位置计数，重试不会倒退或重复累计，成功前不发送完整进度。
func (p *transferProgress) update(value int64) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.finished || p.ctx.Err() != nil || value < p.transferred {
		return
	}
	p.transferred = value
	if p.total >= 0 {
		value = min(value, p.total-1)
	}
	if value <= p.reported || time.Since(p.last) < progressInterval {
		return
	}
	p.reported, p.last = value, time.Now()
	p.callback(value, p.total)
}

// complete 在操作成功后通知最终字节数，未知长度此时转为实际总长度。
func (p *transferProgress) complete(total int64) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.finished || p.ctx.Err() != nil {
		return
	}
	p.finished = true
	p.callback(total, total)
}

// reader 保留 SDK 所需的 Seek 能力，仅读取时统计，不统计 Seek 探测和回退。
func (p *transferProgress) reader(body io.ReadSeeker, offset int64) io.ReadSeeker {
	if p == nil {
		return body
	}
	return &progressReadSeeker{body: body, progress: p, offset: offset}
}

// progressReadSeeker 不暴露 WriterTo，避免 SDK 通过 io.Copy 绕过 Read 计数。
type progressReadSeeker struct {
	body     io.ReadSeeker
	progress *transferProgress
	offset   int64
	position int64
}

// Read 累计 SDK 消费的位置；字节被 SDK 读取不代表云端对象已提交。
func (r *progressReadSeeker) Read(data []byte) (int, error) {
	n, err := r.body.Read(data)
	r.position += int64(n)
	if n > 0 {
		r.progress.update(r.offset + r.position)
	}
	return n, err
}

// Seek 更新当前读取位置，重试产生的重复字节由进度状态过滤。
func (r *progressReadSeeker) Seek(offset int64, whence int) (int64, error) {
	position, err := r.body.Seek(offset, whence)
	if err == nil {
		r.position = position
	}
	return position, err
}

// progressReadCloser 在下载流完整读到 EOF 后通知完成，提前关闭或读取失败不通知完成。
type progressReadCloser struct {
	mu       sync.Mutex
	ctx      context.Context
	body     io.ReadCloser
	progress *transferProgress
	read     int64
	closed   bool
	failed   bool
}

// Read 检查取消和响应长度，不持锁等待底层读取，使 Close 能及时中断网络流。
func (r *progressReadCloser) Read(data []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := r.body.Read(data)
	if cause := r.ctx.Err(); cause != nil {
		err = cause
	}
	r.mu.Lock()
	r.read += int64(n)
	read := r.read
	if err != nil && !errors.Is(err, io.EOF) {
		r.failed = true
	}
	complete := errors.Is(err, io.EOF) && !r.closed && !r.failed
	if complete && r.progress.total >= 0 && read != r.progress.total {
		complete, r.failed, err = false, true, io.ErrUnexpectedEOF
	}
	closed := r.closed
	r.mu.Unlock()
	if !closed {
		r.progress.update(read)
		if complete {
			r.progress.complete(read)
		}
	}
	return n, err
}

// Close 直接关闭底层流，不等待可能阻塞的 Read，也不会将提前关闭视为完成。
func (r *progressReadCloser) Close() error {
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	return r.body.Close()
}

// progressWriter 累计实际落盘字节并保留写入错误，避免把磁盘错误作为网络中断重试。
type progressWriter struct {
	writer   io.Writer
	progress *transferProgress
	written  int64
	err      error
}

// Write 将写入结果交给累计进度，文件发布成功后的完成通知由 DownloadFile 负责。
func (w *progressWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	w.written += int64(n)
	w.err = err
	w.progress.update(w.written)
	return n, err
}
