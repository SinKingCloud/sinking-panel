package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ConnectionInfo 仅提取监听和认证字段，不返回启动参数中的密钥或其他敏感信息。
func (m *Manager) ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error) {
	defer func() { err = m.cleanError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var config bson.M
	if err = m.client.Database("admin").RunCommand(ctx, bson.D{{
		Key:   "getCmdLineOpts",
		Value: 1,
	}}).Decode(&config); err != nil {
		return nil, err
	}
	parsed, _ := config["parsed"].(bson.M)
	network, _ := parsed["net"].(bson.M)
	security, _ := parsed["security"].(bson.M)
	items := make([]map[string]string, 0, 3)
	if value, ok := network["bindIp"]; ok {
		items = append(items, map[string]string{
			"label": "监听地址",
			"value": fmt.Sprint(value),
		})
	} else if all, ok := network["bindIpAll"].(bool); ok && all {
		items = append(items, map[string]string{
			"label": "监听地址",
			"value": "所有地址",
		})
	}
	if value, ok := network["port"]; ok {
		items = append(items, map[string]string{
			"label": "监听端口",
			"value": fmt.Sprint(value),
		})
	}
	if value, ok := security["authorization"]; ok {
		text := fmt.Sprint(value)
		if text == "enabled" {
			text = "已启用"
		} else if text == "disabled" {
			text = "未启用"
		}
		items = append(items, map[string]string{
			"label": "访问认证",
			"value": text,
		})
	}
	return map[string]interface{}{
		"items": items,
		"note":  "账户创建在所选数据库，连接时应将其作为认证库。远程连接还需设置服务器 bindIp、访问认证和端口放行规则。",
	}, nil
}
