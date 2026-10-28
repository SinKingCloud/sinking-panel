package oss

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// Presign 生成临时下载、读取信息或上传地址，调用方需同时携带返回的 Headers。
func (m *Manager) Presign(ctx context.Context, key, method string, ttl time.Duration) (*SignedURL, error) {
	if err := m.objectReady(ctx, key); err != nil {
		return nil, err
	}
	method = strings.ToUpper(method)
	if method != http.MethodGet && method != http.MethodPut && method != http.MethodHead {
		return nil, invalid("预签名仅支持 GET、PUT、HEAD")
	}
	if ttl < time.Second || ttl > 7*24*time.Hour {
		return nil, invalid("签名有效期必须为 1 秒..7 天")
	}
	return m.driver.Presign(ctx, key, method, ttl)
}
