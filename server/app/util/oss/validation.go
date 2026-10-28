package oss

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	bucketNamePattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
	regionNamePattern  = regexp.MustCompile(`^[a-z0-9-]+$`)
	objectRangePattern = regexp.MustCompile(`^bytes=(\d+-\d*|-\d+)$`)
)

// ready 检查调用上下文、实例状态及操作所需的存储桶。
func (m *Manager) ready(ctx context.Context, bucket bool) error {
	if ctx == nil {
		return invalid("context 不能为空")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if m == nil || m.driver == nil {
		return invalid("Manager 未初始化")
	}
	if bucket && m.config.Bucket == "" {
		return invalid("此操作需要 Bucket")
	}
	return nil
}

// validKey 校验对象名称，保留原始路径和特殊字符。
func validKey(key string) error {
	if key == "" || len(key) > 1024 || !utf8.ValidString(key) || strings.ContainsAny(key, "\x00\r\n") {
		return invalid("对象名称需为 1..1024 字节的有效 UTF-8 字符串且不含控制换行")
	}
	// 对象 key 是不透明名称，不使用 filepath.Clean，保留斜杠、空格和点号。
	return nil
}

// objectReady 检查对象操作的公共前置条件。
func (m *Manager) objectReady(ctx context.Context, key string) error {
	if err := m.ready(ctx, true); err != nil {
		return err
	}
	return validKey(key)
}

// pageLimit 应用默认分页大小，并校验厂商共同支持的范围。
func pageLimit(limit int) (int, error) {
	if limit == 0 {
		return 1000, nil
	}
	if limit < 1 || limit > 1000 {
		return 0, invalid("分页大小必须为 1..1000")
	}
	return limit, nil
}
