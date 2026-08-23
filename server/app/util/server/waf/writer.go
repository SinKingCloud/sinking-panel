package waf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/DeRuina/timberjack"
	"github.com/corazawaf/coraza/v3/experimental/plugins"
	"github.com/corazawaf/coraza/v3/experimental/plugins/plugintypes"
)

const writerName = "rotating"

// writer 使用统一的轮转策略写入单站点 WAF 审计日志。
type writer struct {
	mu        sync.Mutex
	formatter plugintypes.AuditLogFormatter
	logger    *timberjack.Logger
}

func init() {
	plugins.RegisterAuditLogWriter(writerName, func() plugintypes.AuditLogWriter {
		return &writer{}
	})
}

func (w *writer) Init(config plugintypes.AuditLogConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if config.Target == "" {
		return errors.New("WAF 审计日志路径不能为空")
	}
	if config.Formatter == nil {
		return errors.New("WAF 审计日志格式不能为空")
	}
	if err := os.MkdirAll(filepath.Dir(config.Target), 0700); err != nil {
		return fmt.Errorf("创建 WAF 审计日志目录失败: %w", err)
	}
	file, err := os.OpenFile(config.Target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("打开 WAF 审计日志失败: %w", err)
	}
	err = errors.Join(file.Chmod(0600), file.Close())
	if err != nil {
		return fmt.Errorf("初始化 WAF 审计日志失败: %w", err)
	}
	w.formatter = config.Formatter
	w.logger = &timberjack.Logger{
		Filename:    config.Target,
		MaxSize:     50,
		MaxBackups:  10,
		MaxAge:      30,
		Compression: "gzip",
		FileMode:    0600,
	}
	return nil
}

func (w *writer) Write(auditLog plugintypes.AuditLog) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.logger == nil || w.formatter == nil {
		return errors.New("WAF 审计日志写入器未初始化")
	}
	data, err := w.formatter.Format(auditLog)
	if err != nil || len(data) == 0 {
		return err
	}
	written, err := w.logger.Write(append(data, '\n'))
	if err == nil && written != len(data)+1 {
		err = io.ErrShortWrite
	}
	return err
}

func (w *writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.logger == nil {
		return nil
	}
	err := w.logger.Close()
	w.logger = nil
	w.formatter = nil
	return err
}
