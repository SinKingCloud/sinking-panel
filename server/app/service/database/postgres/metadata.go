package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schemas 列出 Postgres 业务模式。
func (m *Manager) Schemas(ctx context.Context) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	records, err := m.records(ctx, "SELECT schema_name AS name FROM information_schema.schemata WHERE schema_name NOT IN ('information_schema', 'pg_catalog') AND schema_name NOT LIKE 'pg_toast%' ORDER BY schema_name")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(records))
	for _, row := range records {
		names = append(names, fmt.Sprint(row["name"]))
	}
	return names, nil
}

// Tables 列出表、视图和物化视图。
func (m *Manager) Tables(ctx context.Context, schemaName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	query := "SELECT c.relname AS name, n.nspname AS schema, CASE WHEN c.relkind = 'v' THEN 'VIEW' WHEN c.relkind = 'm' THEN 'MATERIALIZED VIEW' ELSE 'BASE TABLE' END AS type FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind IN ('r','v','m','p','f') AND n.nspname NOT IN ('information_schema','pg_catalog') AND n.nspname NOT LIKE 'pg_toast%'"
	var args []interface{}
	if schemaName != "" {
		query += " AND n.nspname = $1"
		args = append(args, schemaName)
	}
	query += " ORDER BY n.nspname, c.relname"
	return m.records(ctx, query, args...)
}

// Columns 读取 Postgres 字段、主键与只读生成属性。
func (m *Manager) Columns(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	schema, _, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	generated, identity := "false", "''"
	if version >= 100000 {
		identity = "a.attidentity"
	}
	if version >= 120000 {
		generated = "a.attgenerated <> ''"
	}
	query := "SELECT a.attname AS name, format_type(a.atttypid,a.atttypmod) AS type, NOT a.attnotnull AS nullable, pg_get_expr(d.adbin,d.adrelid) AS default_value, " + generated + " AS generated, " + identity + " AS identity, EXISTS (SELECT 1 FROM pg_constraint p WHERE p.conrelid = c.oid AND p.contype = 'p' AND a.attnum = ANY(p.conkey)) AS primary_key FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid JOIN pg_namespace n ON n.oid = c.relnamespace LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = a.attnum WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum"
	args := []interface{}{schema, tableName}
	records, err := m.records(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		record["default"] = record["default_value"]
		delete(record, "default_value")
		record["nullable"] = m.truth(record["nullable"])
		record["primary_key"] = m.truth(record["primary_key"])
		record["generated"] = m.truth(record["generated"])
		record["identity"] = fmt.Sprint(record["identity"])
		// GENERATED ALWAYS 身份列不可写；BY DEFAULT 仍允许显式导入既有值。
		record["generated"] = m.truth(record["generated"]) || record["identity"] == "a"
		record["auto_increment"] = record["identity"] != "" || strings.HasPrefix(fmt.Sprint(record["default"]), "nextval(")
		record["default_expression"] = record["default"] != nil
	}
	return records, nil
}

// Indexes 读取 Postgres 索引及组成表达式。
func (m *Manager) Indexes(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	schema, _, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	query := "SELECT i.relname AS name, pg_get_indexdef(x.indexrelid, k.position + 1, TRUE) AS column_name, x.indisunique AS is_unique, x.indisprimary AS is_primary FROM pg_index x JOIN pg_class t ON t.oid = x.indrelid JOIN pg_namespace n ON n.oid = t.relnamespace JOIN pg_class i ON i.oid = x.indexrelid CROSS JOIN generate_series(0, current_setting('max_index_keys')::int - 1) AS k(position) WHERE n.nspname = $1 AND t.relname = $2 AND k.position < x.indnatts ORDER BY i.relname, k.position"
	args := []interface{}{schema, tableName}
	records, err := m.records(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0)
	positions := make(map[string]int)
	for _, record := range records {
		name := fmt.Sprint(record["name"])
		position, ok := positions[name]
		if !ok {
			position = len(result)
			positions[name] = position
			result = append(result, map[string]interface{}{"name": name, "columns": []string{}, "unique": m.truth(record["is_unique"]), "primary": m.truth(record["is_primary"])})
		}
		column := "(expression)"
		if record["column_name"] != nil {
			column = fmt.Sprint(record["column_name"])
		}
		result[position]["columns"] = append(result[position]["columns"].([]string), column)
	}
	return result, nil
}

func (m *Manager) truth(value interface{}) bool {
	if value == nil {
		return false
	}
	text := strings.ToLower(fmt.Sprint(value))
	if text == "true" || text == "yes" {
		return true
	}
	n, err := strconv.ParseInt(text, 10, 64)
	return err == nil && n != 0
}

func (m *Manager) records(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, 10000)
	if err != nil {
		return nil, err
	}
	if result.HasMore || result.Truncated {
		return nil, errors.New("数据库结构信息超过单次读取上限")
	}
	for index, name := range result.Columns {
		result.Columns[index] = strings.ToLower(name)
	}
	records := make([]map[string]interface{}, 0, len(result.Rows))
	for _, values := range result.Rows {
		record := make(map[string]interface{}, len(result.Columns))
		for index, name := range result.Columns {
			record[name] = values[index]
		}
		records = append(records, record)
	}
	return records, nil
}
