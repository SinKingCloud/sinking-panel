package mysql

import (
	"context"
	"errors"
	"time"
)

// Version 返回服务端版本及当前管理功能的版本能力，权限仍由数据库检查。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.version == "" {
		if err := m.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&m.version); err != nil {
			return nil, m.safeError(err)
		}
	}
	version, maria, err := m.serverVersion(ctx)
	if err != nil {
		return nil, m.safeError(err)
	}
	if version <= 0 {
		return nil, errors.New("无法识别数据库服务器版本")
	}
	name := "MySQL"
	if maria {
		name = "MariaDB"
	}
	return map[string]interface{}{
		"name":    name,
		"version": m.version,
		"features": map[string]bool{
			"accounts":           version >= 50002,
			"views":              version >= 50000,
			"routines":           version >= 50000,
			"functions":          version >= 50000,
			"procedures":         version >= 50000,
			"triggers":           version >= 50002,
			"foreign_keys":       version >= 50110, // 当前外键管理依赖 REFERENTIAL_CONSTRAINTS 目录。
			"default_expression": !maria && version >= 80013 || maria && version >= 100201,
			"column_comments":    version >= 40100,
			"index_comments":     version >= 50503,
			"json_type":          !maria && version >= 50708 || maria && version >= 100207,
			"jsonb_type":         false,
			"materialized_views": false,
		},
	}, nil
}
