package sqlite

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schemas 返回可访问的模式列表。
func (m *Manager) Schemas(ctx context.Context) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []string{"main"}, nil
}

// Tables 返回指定模式的数据表与视图。
func (m *Manager) Tables(ctx context.Context, schemaName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if schemaName != "" && schemaName != "main" {
		return nil, errors.New("Sqlite 仅支持 main 数据库")
	}
	return m.records(ctx, "SELECT name, 'main' AS schema, CASE type WHEN 'view' THEN 'VIEW' ELSE 'BASE TABLE' END AS type FROM main.sqlite_master WHERE type IN ('table','view') AND name NOT GLOB 'sqlite_*' ORDER BY name")
}

// Columns 返回字段、默认值和主键信息。
func (m *Manager) Columns(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := m.target(schemaName, tableName); err != nil {
		return nil, err
	}
	table, _ := m.quote(tableName)
	records, err := m.records(ctx, "PRAGMA main.table_xinfo("+table+")")
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		records, err = m.records(ctx, "PRAGMA main.table_info("+table+")")
		if err != nil {
			return nil, err
		}
	}
	var definition string
	if err = m.db.QueryRowContext(ctx, "SELECT COALESCE(sql, '') FROM main.sqlite_master WHERE name = ? COLLATE NOCASE AND type IN ('table', 'view')", tableName).Scan(&definition); err != nil {
		return nil, err
	}
	// AUTOINCREMENT 仅存在于 INTEGER 单字段主键；普通 INTEGER 主键仍由 rowid 自动赋值。
	keywords, err := m.schemaTokens(definition)
	if err != nil {
		return nil, err
	}
	autoIncrement := false
	for _, keyword := range keywords {
		if keyword == "AUTOINCREMENT" {
			autoIncrement = true
			break
		}
	}
	for _, record := range records {
		record["nullable"] = !m.truth(record["notnull"])
		record["primary_key"] = m.truth(record["pk"])
		record["default"] = record["dflt_value"]
		record["default_expression"] = record["dflt_value"] != nil
		record["auto_increment"] = autoIncrement && m.truth(record["pk"]) && strings.EqualFold(fmt.Sprint(record["type"]), "INTEGER")
		if m.truth(record["auto_increment"]) {
			record["nullable"] = false
		}
		record["generated"] = m.truth(record["hidden"])
		delete(record, "cid")
		delete(record, "pk")
		delete(record, "notnull")
		delete(record, "dflt_value")
		delete(record, "hidden")
	}
	return records, nil
}

// Indexes 返回索引与组成字段。
func (m *Manager) Indexes(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := m.target(schemaName, tableName); err != nil {
		return nil, err
	}
	table, _ := m.quote(tableName)

	records, err := m.records(ctx, "PRAGMA main.index_list("+table+")")
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		name := fmt.Sprint(record["name"])
		quoted, _ := m.quote(name)
		columns, err := m.records(ctx, "PRAGMA main.index_info("+quoted+")")
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(columns))
		for _, column := range columns {
			if column["name"] == nil {
				names = append(names, "(expression)")
			} else {
				names = append(names, fmt.Sprint(column["name"]))
			}
		}
		result = append(result, map[string]interface{}{"name": name, "columns": names, "unique": m.truth(record["unique"]), "primary": record["origin"] == "pk"})
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
