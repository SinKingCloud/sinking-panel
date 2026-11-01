package oss

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode"
)

// Sign 生成临时访问地址，GET 可覆盖响应内容类型和展示方式，调用方需携带返回的 Headers。
func (m *Manager) Sign(ctx context.Context, key, method string, ttl time.Duration, options ...SignOptions) (*SignedURL, error) {
	if err := m.fileReady(ctx, key); err != nil {
		return nil, err
	}
	method = strings.ToUpper(method)
	if method != http.MethodGet && method != http.MethodPut && method != http.MethodHead {
		return nil, m.invalid("预签名仅支持 GET、PUT、HEAD")
	}
	if ttl < time.Second || ttl > 7*24*time.Hour {
		return nil, m.invalid("签名有效期必须为 1 秒..7 天")
	}
	if len(options) > 1 {
		return nil, m.invalid("签名响应配置最多只能提供一项")
	}
	if len(options) == 1 {
		if method != http.MethodGet {
			return nil, m.invalid("只有 GET 签名支持响应配置")
		}
		option := options[0]
		if len(option.ContentType) > 256 || len(option.ContentDisposition) > 4096 ||
			strings.ContainsFunc(option.ContentType, unicode.IsControl) || strings.ContainsFunc(option.ContentDisposition, unicode.IsControl) {
			return nil, m.invalid("签名响应配置过长或包含控制字符")
		}
	}
	return m.driver.Sign(ctx, key, method, ttl, options...)
}
