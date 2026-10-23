package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Version 返回当前运行的 Sqlite 引擎版本，不是数据库文件创建时的版本。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var version string
	if err := m.db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
		return nil, err
	}
	var foreignKeys sql.NullBool
	// 未编译外键支持的 SQLite 不返回这一项；不把读取错误解释为不支持。
	err := m.db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return map[string]interface{}{
		"name":    "Sqlite",
		"version": version,
		"features": map[string]bool{
			"accounts":           false,
			"views":              true,
			"routines":           false,
			"functions":          false,
			"procedures":         false,
			"triggers":           true,
			"foreign_keys":       foreignKeys.Valid,
			"default_expression": true,
			"json_type":          false,
			"jsonb_type":         false,
			"materialized_views": false,
		},
	}, nil
}
