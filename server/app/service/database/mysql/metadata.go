package mysql

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Tables 列出当前数据库的表与视图。
func (m *Manager) Tables(ctx context.Context, schemaName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	if schemaName != "" && schemaName != m.database {
		return nil, errors.New("只能列举当前登记数据库的表")
	}
	return m.records(ctx, "SELECT TABLE_NAME AS name, TABLE_SCHEMA AS `schema`, TABLE_TYPE AS type, ENGINE AS engine FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME", m.database)
}

// Columns 读取字段类型、默认值、主键和生成列属性。
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
	records, err := m.records(ctx, "SELECT COLUMN_NAME AS name, COLUMN_TYPE AS type, IS_NULLABLE AS nullable, COLUMN_DEFAULT AS default_value, CASE WHEN EXISTS (SELECT 1 FROM information_schema.STATISTICS s WHERE s.TABLE_SCHEMA = c.TABLE_SCHEMA AND s.TABLE_NAME = c.TABLE_NAME AND s.COLUMN_NAME = c.COLUMN_NAME AND s.INDEX_NAME = 'PRIMARY') THEN 'PRI' ELSE '' END AS key_kind, EXTRA AS extra, COLUMN_COMMENT AS comment FROM information_schema.COLUMNS c WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION", schema, tableName)
	if err != nil {
		return nil, err
	}
	version, maria, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	mode := ""
	if maria && version >= 100207 {
		// MariaDB 10.2.7 起 COLUMN_DEFAULT 给字符串加 SQL 引号，MySQL 仍返回原值。
		for _, record := range records {
			if strings.Contains(fmt.Sprint(record["default_value"]), "\\") {
				if err = m.db.QueryRowContext(ctx, "SELECT @@SESSION.sql_mode").Scan(&mode); err != nil {
					return nil, err
				}
				break
			}
		}
	}
	noBackslash := strings.Contains(strings.ToUpper(mode), "NO_BACKSLASH_ESCAPES")
	numberDefault := regexp.MustCompile(`(?i)^(?:TRUE|FALSE|[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?)$`)
	for _, record := range records {
		record["default"] = record["default_value"]
		mariaExpression := false
		if maria && version >= 100207 && record["default"] != nil {
			value := fmt.Sprint(record["default"])
			if strings.EqualFold(value, "NULL") {
				record["default"] = nil
			} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
				var literal strings.Builder
				for index := 0; index < len(value); index++ {
					ch := value[index]
					if ch == '\'' && index+1 < len(value) && value[index+1] == '\'' {
						index++
					} else if ch == '\\' && index+1 < len(value) && !noBackslash {
						index++
						ch = value[index]
						switch ch {
						case '0':
							ch = 0
						case 'b':
							ch = '\b'
						case 'n':
							ch = '\n'
						case 'r':
							ch = '\r'
						case 't':
							ch = '\t'
						case 'Z':
							ch = 26
						case '%', '_':
							literal.WriteByte('\\')
						}
					}
					literal.WriteByte(ch)
				}
				record["default"] = literal.String()
			} else {
				mariaExpression = !numberDefault.MatchString(value)
			}
		}
		record["primary_key"] = record["key_kind"] == "PRI"
		record["nullable"] = m.truth(record["nullable"])
		extra := strings.ToUpper(fmt.Sprint(record["extra"]))
		record["generated"] = strings.Contains(extra, "VIRTUAL GENERATED") || strings.Contains(extra, "STORED GENERATED") || strings.Contains(extra, "PERSISTENT GENERATED")
		record["auto_increment"] = strings.Contains(extra, "AUTO_INCREMENT")
		defaultValue := strings.ToUpper(fmt.Sprint(record["default"]))
		record["default_expression"] = record["default"] != nil && (mariaExpression || strings.Contains(extra, "DEFAULT_GENERATED") || (strings.HasPrefix(defaultValue, "CURRENT_TIMESTAMP") && (strings.HasPrefix(strings.ToUpper(fmt.Sprint(record["type"])), "TIMESTAMP") || strings.HasPrefix(strings.ToUpper(fmt.Sprint(record["type"])), "DATETIME"))))
		delete(record, "default_value")
		delete(record, "key_kind")
		delete(record, "extra")
	}
	return records, nil
}

// Indexes 列出 MySQL 索引与组成字段。
func (m *Manager) Indexes(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	// SHOW INDEX 在旧版没有 Index_comment 列，按实际返回字段判断，避免查询不存在的列。
	records, err := m.records(ctx, "SHOW INDEX FROM "+table)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0)
	positions := make(map[string]int)
	for _, record := range records {
		name := fmt.Sprint(record["key_name"])
		position, ok := positions[name]
		if !ok {
			position = len(result)
			positions[name] = position
			comment, supported := record["index_comment"]
			if comment == nil {
				comment = ""
			}
			result = append(result, map[string]interface{}{
				"name":              name,
				"columns":           []string{},
				"unique":            !m.truth(record["non_unique"]),
				"primary":           name == "PRIMARY",
				"comment":           comment,
				"comment_supported": supported,
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
