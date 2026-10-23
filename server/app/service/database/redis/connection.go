package redis

import (
	"context"
	"time"
)

// ConnectionInfo 只读取网络配置，不读取 requirepass 或 ACL 密码摘要。
func (m *Manager) ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error) {
	defer func() { err = m.cleanError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	items := make([]map[string]string, 0, 3)
	for _, field := range []struct{ key, label string }{{"bind", "监听地址"}, {"port", "监听端口"}, {"protected-mode", "保护模式"}} {
		// 旧版本 CONFIG GET 一次只支持一个参数；不存在的配置不推测其值。
		values, err := m.client.ConfigGet(ctx, field.key).Result()
		if err != nil {
			return nil, err
		}
		if value, ok := values[field.key]; ok {
			if field.key == "bind" && value == "" {
				value = "所有地址"
			} else if field.key == "protected-mode" {
				if value == "yes" {
					value = "已启用"
				} else if value == "no" {
					value = "未启用"
				}
			}
			items = append(items, map[string]string{"label": field.label, "value": value})
		}
	}
	return map[string]interface{}{"items": items, "note": "ACL 账户作用于整个 Redis 实例。远程连接还需服务器监听对应地址、允许端口访问并设置认证。"}, nil
}
