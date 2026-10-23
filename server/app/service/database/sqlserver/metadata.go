package sqlserver

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schemas 返回可访问的模式列表。
func (m *Manager) Schemas(ctx context.Context) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	query := "SELECT name FROM sys.schemas WHERE name NOT IN ('sys', 'INFORMATION_SCHEMA') ORDER BY name"
	records, err := m.records(ctx, query)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(records))
	for _, row := range records {
		names = append(names, fmt.Sprint(row["name"]))
	}
	return names, nil
}

// Tables 返回指定模式的数据表与视图。
func (m *Manager) Tables(ctx context.Context, schemaName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// 局部临时表和表变量仅属于创建它们的会话，不能通过后续请求打开；保留全局临时表。
	query := "SELECT TABLE_NAME AS name, TABLE_SCHEMA AS [schema], TABLE_TYPE AS type FROM INFORMATION_SCHEMA.TABLES WHERE (TABLE_NAME NOT LIKE '#%' OR TABLE_NAME LIKE '##%')"
	var args []interface{}
	if schemaName != "" {
		query += " AND TABLE_SCHEMA = @p1"
		args = append(args, schemaName)
	}
	query += " ORDER BY TABLE_SCHEMA, TABLE_NAME"
	return m.records(ctx, query, args...)
}

// Columns 返回字段、默认值和主键信息。
func (m *Manager) Columns(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
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
	generatedKind := "0"
	if version >= 13 {
		generatedKind = "c.generated_always_type"
	}
	query := "SELECT c.name, CASE WHEN t.is_user_defined = 1 THEN QUOTENAME(SCHEMA_NAME(t.schema_id)) + '.' + QUOTENAME(t.name) WHEN t.name IN ('varchar', 'char', 'varbinary', 'binary') THEN t.name + '(' + CASE WHEN c.max_length = -1 THEN 'max' ELSE CAST(c.max_length AS VARCHAR(10)) END + ')' WHEN t.name IN ('nvarchar', 'nchar') THEN t.name + '(' + CASE WHEN c.max_length = -1 THEN 'max' ELSE CAST(c.max_length / 2 AS VARCHAR(10)) END + ')' WHEN t.name IN ('decimal', 'numeric') THEN t.name + '(' + CAST(c.precision AS VARCHAR(10)) + ',' + CAST(c.scale AS VARCHAR(10)) + ')' WHEN t.name IN ('time', 'datetime2', 'datetimeoffset') THEN t.name + '(' + CAST(c.scale AS VARCHAR(10)) + ')' ELSE t.name END AS type, c.is_nullable AS nullable, c.collation_name AS collation, d.definition AS default_value, c.is_computed AS generated, c.is_computed AS computed, " + generatedKind + " AS generated_kind, c.is_identity AS identity_column, CASE WHEN EXISTS (SELECT 1 FROM sys.indexes i JOIN sys.index_columns ic ON ic.object_id = i.object_id AND ic.index_id = i.index_id WHERE i.object_id = c.object_id AND i.is_primary_key = 1 AND ic.column_id = c.column_id) THEN 1 ELSE 0 END AS primary_key, TYPE_NAME(c.system_type_id) AS base_type FROM sys.columns c JOIN sys.objects o ON o.object_id = c.object_id JOIN sys.schemas s ON s.schema_id = o.schema_id JOIN sys.types t ON t.user_type_id = c.user_type_id LEFT JOIN sys.default_constraints d ON d.object_id = c.default_object_id WHERE s.name = @p1 AND o.name = @p2 ORDER BY c.column_id"
	records, err := m.records(ctx, query, schema, tableName)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		record["default"] = record["default_value"]
		record["default_expression"] = record["default_value"] != nil
		delete(record, "default_value")
		record["nullable"] = m.truth(record["nullable"])
		record["primary_key"] = m.truth(record["primary_key"])
		record["computed"] = m.truth(record["computed"])
		record["generated"] = m.truth(record["generated"]) || m.truth(record["generated_kind"]) || m.truth(record["identity_column"]) || strings.EqualFold(fmt.Sprint(record["type"]), "timestamp") || strings.EqualFold(fmt.Sprint(record["type"]), "rowversion")
		record["auto_increment"] = m.truth(record["identity_column"])
	}
	return records, nil
}

// Indexes 返回索引与组成字段。
func (m *Manager) Indexes(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	schema, _, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}

	query := "SELECT i.name, c.name AS column_name, i.is_unique, i.is_primary_key AS is_primary FROM sys.indexes i JOIN sys.objects o ON o.object_id = i.object_id JOIN sys.schemas s ON s.schema_id = o.schema_id JOIN sys.index_columns ic ON ic.object_id = i.object_id AND ic.index_id = i.index_id JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id WHERE s.name = @p1 AND o.name = @p2 AND i.name IS NOT NULL AND ic.is_included_column = 0 ORDER BY i.name, ic.key_ordinal"
	records, err := m.records(ctx, query, schema, tableName)
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
			result = append(result, map[string]interface{}{
				"name": name, "columns": []string{}, "unique": m.truth(record["is_unique"]), "primary": m.truth(record["is_primary"]),
			})
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
