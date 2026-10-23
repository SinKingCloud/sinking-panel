package postgres

import (
	"context"
	"errors"
	"time"
)

// Version 返回 Postgres 实际版本，单独标识存储过程与物化视图等版本差异。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	version, err := m.serverVersion(ctx)
	if err != nil {
		return nil, m.safeError(err)
	}
	if version <= 0 {
		return nil, errors.New("无法识别数据库服务器版本")
	}
	var text string
	if err = m.db.QueryRowContext(ctx, "SHOW server_version").Scan(&text); err != nil {
		return nil, m.safeError(err)
	}
	return map[string]interface{}{
		"name":    "Postgres",
		"version": text,
		"features": map[string]bool{
			"accounts":           version >= 80100,
			"views":              true,
			"routines":           version >= 80400,
			"functions":          version >= 80400,
			"procedures":         version >= 110000,
			"triggers":           true,
			"foreign_keys":       version >= 80400,
			"default_expression": true,
			"json_type":          version >= 90200,
			"jsonb_type":         version >= 90400,
			"uuid_type":          version >= 80300,
			"materialized_views": version >= 90300,
			"smallserial":        version >= 90200,
		},
	}, nil
}
