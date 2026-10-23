package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ConnectionInfo 读取实际网络配置，不把账号来源授权误认为服务器已开放公网访问。
func (m *Manager) ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	records, err := m.records(ctx, "SHOW VARIABLES WHERE Variable_name IN ('bind_address', 'port', 'skip_networking')")
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	for _, record := range records {
		values[strings.ToLower(fmt.Sprint(record["variable_name"]))] = fmt.Sprint(record["value"])
	}
	items := make([]map[string]string, 0, 3)
	for _, field := range []struct{ key, label string }{{"bind_address", "监听地址"}, {"port", "监听端口"}, {"skip_networking", "TCP 连接"}} {
		if value, ok := values[field.key]; ok {
			if field.key == "skip_networking" {
				if value == "1" || strings.EqualFold(value, "ON") {
					value = "已禁用"
				} else if value == "0" || strings.EqualFold(value, "OFF") {
					value = "已启用"
				}
			}
			items = append(items, map[string]string{"label": field.label, "value": value})
		}
	}
	return map[string]interface{}{"items": items, "note": "连接来源控制账号允许从哪里登录；远程连接还需要服务器监听目标地址，并放行对应端口。"}, nil
}
