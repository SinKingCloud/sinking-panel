package sqlserver

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Version 返回 SqlServer 实际版本，原生类型按当前数据库目录判断，兼容 Azure SQL。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var text string
	if err := m.db.QueryRowContext(ctx, "SELECT CAST(SERVERPROPERTY('ProductVersion') AS VARCHAR(100))").Scan(&text); err != nil {
		return nil, m.safeError(err)
	}
	version := m.version
	if version == 0 {
		var err error
		version, err = strconv.Atoi(strings.Split(text, ".")[0])
		if err != nil {
			return nil, m.safeError(err)
		}
	}
	if version <= 0 {
		return nil, errors.New("无法识别数据库服务器版本")
	}
	var jsonType bool
	if version >= 9 {
		// Azure SQL 的产品主版本不能代表功能版本；只检查内置类型，不接受同名自定义类型。
		if err := m.db.QueryRowContext(ctx, "SELECT CASE WHEN EXISTS (SELECT 1 FROM sys.types WHERE name = 'json' AND is_user_defined = 0) THEN 1 ELSE 0 END").Scan(&jsonType); err != nil {
			return nil, m.safeError(err)
		}
	}
	return map[string]interface{}{
		"name":    "SqlServer",
		"version": text,
		"features": map[string]bool{
			"accounts":           version >= 9,
			"views":              version >= 9,
			"routines":           version >= 9,
			"functions":          version >= 9,
			"procedures":         version >= 9,
			"triggers":           version >= 9,
			"foreign_keys":       version >= 9,
			"default_expression": true,
			"json_type":          jsonType,
			"jsonb_type":         false,
			"date_type":          version >= 10,
			"datetime2_type":     version >= 10,
			"materialized_views": false,
		},
	}, nil
}
