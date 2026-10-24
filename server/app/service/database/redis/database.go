package redis

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// DatabaseInfo 读取当前逻辑库的键数量及过期统计，不返回其他逻辑库的数据。
func (m *Manager) DatabaseInfo(ctx context.Context) (result map[string]interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	keys, err := m.client.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}
	result = map[string]interface{}{"name": strconv.Itoa(m.database), "exists": true, "keys": keys}
	info, infoErr := m.info(ctx, "keyspace")
	if infoErr != nil {
		// INFO 可能被 ACL 禁用，DBSIZE 成功时仍保留实际键数量。
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return result, nil
	}
	if keys == 0 {
		result["expires"], result["avg_ttl"] = int64(0), int64(0)
	}
	for _, line := range strings.Split(info, "\n") {
		name, content, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok || name != "db"+strconv.Itoa(m.database) {
			continue
		}
		for _, field := range strings.Split(content, ",") {
			key, value, ok := strings.Cut(field, "=")
			if !ok || (key != "expires" && key != "avg_ttl") {
				continue
			}
			if number, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil && number >= 0 {
				result[key] = number
			}
		}
		break
	}
	return result, nil
}
