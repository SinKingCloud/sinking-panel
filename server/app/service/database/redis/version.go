package redis

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// Version 返回服务器版本及版本支持的功能，不将账户权限视为版本能力。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	info, err := m.info(ctx, "server")
	if err != nil {
		return nil, m.cleanError(err)
	}
	var version string
	for _, line := range strings.Split(info, "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok && key == "redis_version" {
			version = strings.TrimSpace(value)
			break
		}
	}
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return nil, errors.New("无法识别 Redis 服务器版本")
	}
	major, majorErr := strconv.Atoi(parts[0])
	minor, minorErr := strconv.Atoi(parts[1])
	if majorErr != nil || minorErr != nil || major < 1 || minor < 0 {
		return nil, errors.New("无法识别 Redis 服务器版本")
	}
	return map[string]interface{}{
		"name":    "Redis",
		"version": version,
		"features": map[string]bool{
			"accounts": major >= 6, // Redis 6.0 引入 ACL 账户管理。
			"scan":     major > 2 || major == 2 && minor >= 8,
			"streams":  major >= 5,
			"lua":      major > 2 || major == 2 && minor >= 6,
		},
	}, nil
}
