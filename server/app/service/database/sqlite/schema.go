package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// 字段批量建表时复用校验表达式，避免为每个字段重新编译。
var columnTypePattern = regexp.MustCompile(`(?i)^[a-z][a-z0-9_]*(?:\s+[a-z][a-z0-9_]*)*(?:\s*\(\s*\d+\s*(?:,\s*\d+\s*)?\))?$`)

// columnDefinition 只生成单个字段定义，不接受混入约束或其他语句的类型。
func (m *Manager) columnDefinition(column ColumnDefinition) (string, error) {
	name, err := m.quote(column.Name)
	if err != nil {
		return "", err
	}
	typeName := strings.TrimSpace(column.Type)
	if len(typeName) > 128 || !columnTypePattern.MatchString(typeName) {
		return "", errors.New("字段类型不合法，请只填写类型和长度")
	}
	for _, word := range strings.Fields(strings.ToUpper(typeName)) {
		switch word {
		case "PRIMARY", "KEY", "DEFAULT", "REFERENCES", "CHECK", "CONSTRAINT", "UNIQUE", "COLLATE", "GENERATED", "NOT", "NULL", "AUTOINCREMENT":
			return "", errors.New("字段类型中不能包含字段约束")
		}
	}
	definition := name + " " + typeName
	if column.AutoIncrement {
		if !strings.EqualFold(typeName, "INTEGER") || column.Nullable || column.Default != nil {
			return "", errors.New("Sqlite 自增字段必须为 INTEGER 主键、非空且没有默认值")
		}
		return definition + " PRIMARY KEY AUTOINCREMENT", nil
	}
	if !column.Nullable {
		definition += " NOT NULL"
	}
	if column.Default != nil {
		value := *column.Default
		if len(value) > 64<<10 || strings.ContainsRune(value, 0) {
			return "", errors.New("字段默认值不合法或超过 64 KB")
		}
		if column.DefaultExpression {
			if strings.TrimSpace(value) == "" {
				return "", errors.New("默认值表达式必须是单个 SQL 表达式")
			}
			tokens, err := m.schemaTokens(value)
			if err != nil {
				return "", err
			}
			depth := 0
			for _, token := range tokens {
				switch token {
				case "(":
					depth++
				case ")":
					depth--
				case "SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "CREATE", "ALTER", "DROP", "EXEC", "EXECUTE", "USE", "SET", "GRANT", "REVOKE", "DENY", ";":
					return "", errors.New("默认值表达式不能包含 SQL 语句")
				}
				if depth < 0 {
					return "", errors.New("默认值表达式括号不匹配")
				}
			}
			if depth != 0 {
				return "", errors.New("默认值表达式括号不匹配")
			}
			definition += " DEFAULT (" + value + ")"
		} else {
			definition += " DEFAULT '" + strings.ReplaceAll(value, "'", "''") + "'"
		}
	}
	return definition, nil
}

// CreateTable 创建真实数据表，支持复合主键与 INTEGER 自增主键。
func (m *Manager) CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.ToLower(tableName), "sqlite_") {
		return nil, errors.New("不能创建 Sqlite 系统表")
	}
	if len(columns) == 0 || len(columns) > 2000 {
		return nil, errors.New("字段数量必须在 1 到 2000 之间")
	}
	keys := make(map[string]bool, len(primaryKeys))
	for _, key := range primaryKeys {
		lower := strings.ToLower(key)
		if keys[lower] {
			return nil, errors.New("主键字段不能重复")
		}
		keys[lower] = true
	}
	seen := make(map[string]bool, len(columns))
	definitions := make([]string, 0, len(columns)+1)
	autoColumn := ""
	for _, column := range columns {
		lower := strings.ToLower(column.Name)
		if seen[lower] {
			return nil, errors.New("字段名称不能重复")
		}
		seen[lower] = true
		if keys[lower] {
			column.Nullable = false
		}
		if column.AutoIncrement {
			if autoColumn != "" || (len(primaryKeys) > 0 && (len(primaryKeys) != 1 || !keys[lower])) {
				return nil, errors.New("Sqlite 自增字段必须是唯一的单字段主键")
			}
			autoColumn = column.Name
			column.Nullable = false
		}
		definition, err := m.columnDefinition(column)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	quotedKeys := make([]string, 0, len(primaryKeys))
	for _, key := range primaryKeys {
		if !seen[strings.ToLower(key)] {
			return nil, errors.New("主键必须选择已定义的字段")
		}
		quoted, _ := m.quote(key)
		quotedKeys = append(quotedKeys, quoted)
	}
	if len(quotedKeys) > 0 && autoColumn == "" {
		definitions = append(definitions, "PRIMARY KEY ("+strings.Join(quotedKeys, ", ")+")")
	}
	for _, key := range foreignKeys {
		for _, field := range key.Columns {
			if !seen[strings.ToLower(field)] {
				return nil, errors.New("外键必须选择已定义的字段")
			}
		}
		definition, err := m.foreignKeyDefinition(key)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	if len(indexes) > 64 {
		return nil, errors.New("索引数量不能超过 64 个")
	}
	statements := []string{"CREATE TABLE " + table + " (\n  " + strings.Join(definitions, ",\n  ") + "\n)"}
	seenIndexes := make(map[string]bool, len(indexes))
	for _, index := range indexes {
		quotedIndex, quoteErr := m.quote(index.Name)
		if quoteErr != nil {
			return nil, quoteErr
		}
		if seenIndexes[strings.ToLower(index.Name)] || len(index.Columns) == 0 {
			return nil, errors.New("索引名称不能重复且必须选择字段")
		}
		seenIndexes[strings.ToLower(index.Name)] = true
		indexFields := make([]string, 0, len(index.Columns))
		used := make(map[string]bool, len(index.Columns))
		for _, field := range index.Columns {
			found := false
			for _, column := range columns {
				if strings.EqualFold(column.Name, field.Name) {
					found = true
					break
				}
			}
			if !found || used[strings.ToLower(field.Name)] {
				return nil, errors.New("索引字段不存在或重复")
			}
			used[strings.ToLower(field.Name)] = true
			quotedField, quoteErr := m.quote(field.Name)
			if quoteErr != nil {
				return nil, quoteErr
			}
			if field.Desc {
				quotedField += " DESC"
			} else {
				quotedField += " ASC"
			}
			indexFields = append(indexFields, quotedField)
		}
		prefix := "CREATE INDEX "
		if index.Unique {
			prefix = "CREATE UNIQUE INDEX "
		}
		quotedTable, _ := m.quote(tableName)
		statements = append(statements, prefix+`"main".`+quotedIndex+" ON "+quotedTable+" ("+strings.Join(indexFields, ", ")+")")
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	for _, statement := range statements {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// RenameTable 使用原生改名，索引、触发器和外键由 Sqlite 一并维护。
func (m *Manager) RenameTable(ctx context.Context, schemaName, tableName, newName string) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	name, err := m.quote(newName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.ToLower(newName), "sqlite_") {
		return nil, errors.New("不能使用 Sqlite 系统表名称")
	}
	if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" RENAME TO "+name); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// AddColumn 添加字段，依赖引擎校验已有数据与默认值，不重建数据表。
func (m *Manager) AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if column.AutoIncrement {
		return nil, errors.New("Sqlite 不能向已有表添加自增主键")
	}
	definition, err := m.columnDefinition(column)
	if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN "+definition); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// UpdateColumn 单纯改名走原生语法，类型/默认值/非空变更采用事务迁移并保留依赖。
func (m *Manager) UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	name, err := m.quote(column.Name)
	if err != nil {
		return nil, err
	}
	oldName, err := m.quote(columnName)
	if err != nil {
		return nil, err
	}
	data, err := m.Columns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	var current map[string]interface{}
	for _, record := range data.([]map[string]interface{}) {
		if strings.EqualFold(fmt.Sprint(record["name"]), columnName) {
			current = record
			break
		}
	}
	if current == nil {
		return nil, errors.New("字段不存在")
	}
	defaultSame := current["default"] == nil && column.Default == nil
	if current["default"] != nil && column.Default != nil {
		value := *column.Default
		if !column.DefaultExpression {
			value = "'" + strings.ReplaceAll(value, "'", "''") + "'"
		}
		defaultSame = strings.TrimSpace(value) == strings.TrimSpace(fmt.Sprint(current["default"]))
	}
	if column.AutoIncrement != m.truth(current["auto_increment"]) {
		return nil, errors.New("不能隐式添加或移除自增主键机制，请使用显式迁移")
	}
	if !strings.EqualFold(strings.TrimSpace(column.Type), strings.TrimSpace(fmt.Sprint(current["type"]))) || column.Nullable != m.truth(current["nullable"]) || !defaultSame {
		return m.rebuildTable(ctx, schemaName, tableName, &column, columnName, nil, false)
	}
	if column.Name != fmt.Sprint(current["name"]) {
		if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" RENAME COLUMN "+oldName+" TO "+name); err != nil {
			return nil, err
		}
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// DropColumn 使用原生删除；字段被索引、外键或触发器引用时由引擎拒绝。
func (m *Manager) DropColumn(ctx context.Context, schemaName, tableName, columnName string) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	column, err := m.quote(columnName)
	if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" DROP COLUMN "+column); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// CreateIndex 创建普通或唯一索引；已有表的主键不能通过索引接口替换。
func (m *Manager) CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	if primary {
		return nil, errors.New("Sqlite 主键只能在建表时设置，已有表添加主键需要显式迁移")
	}
	if _, err := m.maintenanceTable(ctx, schemaName, tableName); err != nil {
		return nil, err
	}
	if len(columns) == 0 || len(columns) > 2000 {
		return nil, errors.New("请选择索引字段")
	}
	if strings.HasPrefix(strings.ToLower(indexName), "sqlite_") {
		return nil, errors.New("不能使用 Sqlite 系统索引名称")
	}
	index, err := m.quote(indexName)
	if err != nil {
		return nil, err
	}
	table, _ := m.quote(tableName)
	fields := make([]string, 0, len(columns))
	seen := make(map[string]bool, len(columns))
	for _, column := range columns {
		name, err := m.quote(column.Name)
		if err != nil {
			return nil, err
		}
		lower := strings.ToLower(column.Name)
		if seen[lower] {
			return nil, errors.New("索引字段不能重复")
		}
		seen[lower] = true
		if column.Desc {
			name += " DESC"
		} else {
			name += " ASC"
		}
		fields = append(fields, name)
	}
	prefix := "CREATE INDEX "
	if unique {
		prefix = "CREATE UNIQUE INDEX "
	}
	if _, err = m.db.ExecContext(ctx, prefix+`"main".`+index+" ON "+table+" ("+strings.Join(fields, ", ")+")"); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// DropIndex 确认索引属于当前表，只删除显式创建的索引。
func (m *Manager) DropIndex(ctx context.Context, schemaName, tableName, indexName string) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	if _, err := m.maintenanceTable(ctx, schemaName, tableName); err != nil {
		return nil, err
	}
	index, err := m.quote(indexName)
	if err != nil {
		return nil, err
	}
	var statement sql.NullString
	err = m.db.QueryRowContext(ctx, "SELECT sql FROM main.sqlite_master WHERE type = 'index' AND name = ? COLLATE NOCASE AND tbl_name = ? COLLATE NOCASE", indexName, tableName).Scan(&statement)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("当前表中不存在该索引")
	}
	if err != nil {
		return nil, err
	}
	if !statement.Valid || strings.HasPrefix(strings.ToLower(indexName), "sqlite_") {
		return nil, errors.New("不能直接删除主键或唯一约束生成的系统索引")
	}
	if _, err = m.db.ExecContext(ctx, `DROP INDEX "main".`+index); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// DDL 返回真实建表或视图语句，并保留表的索引和触发器定义。
func (m *Manager) DDL(ctx context.Context, schemaName, tableName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err := m.target(schemaName, tableName); err != nil {
		return "", err
	}
	rows, err := m.db.QueryContext(ctx, "SELECT sql FROM main.sqlite_master WHERE sql IS NOT NULL AND ((name = ? COLLATE NOCASE AND type IN ('table', 'view')) OR (tbl_name = ? COLLATE NOCASE AND type IN ('index', 'trigger'))) ORDER BY CASE type WHEN 'table' THEN 0 WHEN 'view' THEN 0 WHEN 'index' THEN 1 ELSE 2 END, name", tableName, tableName)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var script strings.Builder
	for rows.Next() {
		var statement string
		if err = rows.Scan(&statement); err != nil {
			return "", err
		}
		if script.Len()+len(statement) > 1<<20 {
			return "", errors.New("结构定义超过 1 MB")
		}
		script.WriteString(strings.TrimRight(statement, "; \r\n\t"))
		script.WriteString(";\n\n")
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	if script.Len() == 0 {
		return "", errors.New("数据表或视图不存在")
	}
	return strings.TrimSpace(script.String()), nil
}

// CreateView 创建视图，查询内容由数据库解析，不执行独立的附加语句。
func (m *Manager) CreateView(ctx context.Context, schemaName, viewName, statement string) (*Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	view, err := m.target(schemaName, viewName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.ToLower(viewName), "sqlite_") {
		return nil, errors.New("不能使用 Sqlite 系统对象名称")
	}
	statement = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(statement), ";"))
	if statement == "" || len(statement) > 1<<20 || strings.ContainsRune(statement, 0) {
		return nil, errors.New("视图内容必须是单条查询，且不超过 1 MB")
	}
	tokens, err := m.schemaTokens(statement)
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		if token == ";" {
			return nil, errors.New("视图定义不能包含附加语句")
		}
	}
	if _, err = m.db.ExecContext(ctx, "CREATE VIEW "+view+" AS "+statement); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}
