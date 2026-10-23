package postgres

import (
	"context"
	"fmt"
	"time"
)

// ConnectionInfo 读取监听和 TLS 设置；客户端不改写服务器的 pg_hba.conf。
func (m *Manager) ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	records, err := m.records(ctx, "SELECT name, setting FROM pg_settings WHERE name IN ('listen_addresses', 'port', 'ssl')")
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	hidden := false
	for _, record := range records {
		// 部分版本会对无权限的连接隐藏 setting，不把 NULL 当作实际配置值。
		if record["setting"] == nil {
			hidden = true
			continue
		}
		values[fmt.Sprint(record["name"])] = fmt.Sprint(record["setting"])
	}
	items := make([]map[string]string, 0, 3)
	for _, field := range []struct{ key, label string }{{"listen_addresses", "监听地址"}, {"port", "监听端口"}, {"ssl", "TLS 连接"}} {
		if value, ok := values[field.key]; ok {
			if field.key == "ssl" {
				if value == "on" {
					value = "已启用"
				} else if value == "off" {
					value = "未启用"
				}
			}
			items = append(items, map[string]string{"label": field.label, "value": value})
		}
	}
	note := "允许远程登录需在数据库服务器设置 listen_addresses、pg_hba.conf 和端口放行规则；账户权限只控制可访问的数据。"
	if hidden {
		note += "当前账户无权读取部分监听配置。"
	}
	return map[string]interface{}{"items": items, "note": note}, nil
}
