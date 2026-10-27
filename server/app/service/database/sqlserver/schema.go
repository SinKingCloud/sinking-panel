package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// 字段批量建表时复用校验表达式，避免为每个字段重新编译。
var columnTypePattern = regexp.MustCompile(`(?i)^(?:[a-z][a-z0-9_]*|\[(?:[^\]]|\]\])+\]\s*\.\s*\[(?:[^\]]|\]\])+\])(?:\s*\(\s*(?:max|\d+)\s*(?:,\s*\d+\s*)?\))?$`)
var integerDecimalPattern = regexp.MustCompile(`(?i)^(decimal|numeric)\s*\(\s*\d+\s*,\s*0\s*\)$`)

// columnDefinition 生成单字段定义，类型不接受附加约束或独立语句。
func (m *Manager) columnDefinition(column ColumnDefinition) (string, error) {
	name, err := m.quote(column.Name)
	if err != nil || utf8.RuneCountInString(column.Name) > 128 {
		return "", errors.New("字段名称不合法，最多允许 128 个字符")
	}
	typeName := strings.TrimSpace(column.Type)
	if len(typeName) > 256 || !columnTypePattern.MatchString(typeName) {
		return "", errors.New("字段类型不合法，请只填写类型和长度")
	}
	definition := name + " " + typeName
	if column.AutoIncrement {
		base := strings.ToLower(strings.Split(typeName, "(")[0])
		if base != "tinyint" && base != "smallint" && base != "int" && base != "bigint" && !integerDecimalPattern.MatchString(typeName) {
			return "", errors.New("自增字段必须为整数或小数位为 0 的 decimal/numeric")
		}
		if column.Nullable || column.Default != nil {
			return "", errors.New("自增字段必须非空且没有默认值")
		}
		definition += " IDENTITY(1,1)"
	}
	if column.Nullable {
		definition += " NULL"
	} else {
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
			definition += " DEFAULT N'" + strings.ReplaceAll(value, "'", "''") + "'"
		}
	}
	return definition, nil
}

// CreateTable 创建用户表，主键字段自动设为非空。
func (m *Manager) CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(tableName) > 128 || strings.EqualFold(schema, "sys") || strings.EqualFold(schema, "INFORMATION_SCHEMA") {
		return nil, errors.New("不能使用系统模式或超过 128 字符的表名称")
	}
	if len(columns) == 0 || len(columns) > 1024 {
		return nil, errors.New("字段数量必须在 1 到 1024 之间")
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
	autoCount := 0
	for _, column := range columns {
		lower := strings.ToLower(column.Name)
		if seen[lower] {
			return nil, errors.New("字段名称不能重复")
		}
		seen[lower] = true
		if keys[lower] || column.AutoIncrement {
			column.Nullable = false
		}
		if column.AutoIncrement {
			autoCount++
			if autoCount > 1 {
				return nil, errors.New("每张表只能有一个自增字段")
			}
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
	if len(quotedKeys) > 0 {
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
		statements = append(statements, prefix+quotedIndex+" ON "+table+" ("+strings.Join(indexFields, ", ")+")")
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
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// RenameTable 使用绑定参数调用 sp_rename，仅修改当前用户表的名称。
func (m *Manager) RenameTable(ctx context.Context, schemaName, tableName, newName string) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if _, err = m.quote(newName); err != nil || utf8.RuneCountInString(newName) > 128 {
		return nil, errors.New("新表名称不合法，最多允许 128 个字符")
	}
	if _, err = m.db.ExecContext(ctx, "EXEC sys.sp_rename @objname = @p1, @newname = @p2, @objtype = 'OBJECT'", table, newName); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// AddColumn 添加字段，已有数据的非空约束由 SqlServer 原生校验。
func (m *Manager) AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	definition, err := m.columnDefinition(column)
	if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD "+definition); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// UpdateColumn 在同一事务内修改字段和默认约束；不允许增删 IDENTITY 属性。
func (m *Manager) UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	name, err := m.quote(column.Name)
	if err != nil || utf8.RuneCountInString(column.Name) > 128 {
		return nil, errors.New("字段名称不合法，最多允许 128 个字符")
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
		return nil, errors.New("字段不存在或当前账户无权访问")
	}
	identity := m.truth(current["auto_increment"])
	if column.AutoIncrement != identity {
		return nil, errors.New("不能在已有字段上添加或移除 IDENTITY 属性")
	}
	structureChanged := !strings.EqualFold(strings.TrimSpace(column.Type), strings.TrimSpace(fmt.Sprint(current["type"]))) || column.Nullable != m.truth(current["nullable"])
	if m.truth(current["generated"]) && !identity && (structureChanged || column.Default != nil) {
		return nil, errors.New("数据库生成字段仅支持改名，请使用 SQL 显式修改其生成属性")
	}
	if identity && structureChanged {
		return nil, errors.New("自增字段仅支持改名，类型变更请使用 SQL 显式执行")
	}
	definition, err := m.columnDefinition(column)
	if err != nil {
		return nil, err
	}
	defaultSQL := ""
	if position := strings.Index(definition, " DEFAULT "); position >= 0 {
		defaultSQL = definition[position+9:]
	}
	defaultSame := current["default"] == nil && column.Default == nil
	if current["default"] != nil && column.Default != nil {
		defaultSame = column.DefaultExpression && strings.TrimSpace(*column.Default) == strings.TrimSpace(fmt.Sprint(current["default"]))
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if structureChanged || !defaultSame {
		var constraint sql.NullString
		if err = tx.QueryRowContext(ctx, "SELECT d.name FROM sys.columns c LEFT JOIN sys.default_constraints d ON d.object_id = c.default_object_id WHERE c.object_id = OBJECT_ID(@p1) AND c.name = @p2", table, columnName).Scan(&constraint); err != nil {
			return nil, err
		}
		if constraint.Valid {
			quoted, err := m.quote(constraint.String)
			if err != nil {
				return nil, err
			}
			if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+quoted); err != nil {
				return nil, err
			}
		}
	}
	if structureChanged {
		nullable := " NOT NULL"
		if column.Nullable {
			nullable = " NULL"
		}
		collation := ""
		if current["collation"] != nil {
			quoted, err := m.quote(fmt.Sprint(current["collation"]))
			if err != nil {
				return nil, err
			}
			collation = " COLLATE " + quoted
		}
		if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" "+strings.TrimSpace(column.Type)+collation+nullable); err != nil {
			return nil, err
		}
	}
	if column.Name != fmt.Sprint(current["name"]) {
		if _, err = tx.ExecContext(ctx, "EXEC sys.sp_rename @objname = @p1, @newname = @p2, @objtype = 'COLUMN'", table+"."+oldName, column.Name); err != nil {
			return nil, err
		}
	}
	if (structureChanged || !defaultSame) && defaultSQL != "" {
		if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" ADD DEFAULT "+defaultSQL+" FOR "+name); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// DropColumn 同一事务删除字段的默认约束和字段，其他依赖由引擎保护。
func (m *Manager) DropColumn(ctx context.Context, schemaName, tableName, columnName string) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
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
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var constraint sql.NullString
	if err = tx.QueryRowContext(ctx, "SELECT d.name FROM sys.columns c LEFT JOIN sys.default_constraints d ON d.object_id = c.default_object_id WHERE c.object_id = OBJECT_ID(@p1) AND c.name = @p2", table, columnName).Scan(&constraint); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("字段不存在")
	} else if err != nil {
		return nil, err
	}
	if constraint.Valid {
		quoted, _ := m.quote(constraint.String)
		if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+quoted); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" DROP COLUMN "+column); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// CreateIndex 创建普通、唯一索引或主键约束。
func (m *Manager) CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	index, err := m.quote(indexName)
	if err != nil || utf8.RuneCountInString(indexName) > 128 {
		return nil, errors.New("索引名称不合法，最多允许 128 个字符")
	}
	if len(columns) == 0 || len(columns) > 32 {
		return nil, errors.New("索引字段数量必须在 1 到 32 之间")
	}
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
	query := "CREATE "
	if unique {
		query += "UNIQUE "
	}
	query += "INDEX " + index + " ON " + table + " (" + strings.Join(fields, ", ") + ")"
	if primary {
		query = "ALTER TABLE " + table + " ADD CONSTRAINT " + index + " PRIMARY KEY (" + strings.Join(fields, ", ") + ")"
	}
	if _, err = m.db.ExecContext(ctx, query); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// DropIndex 主键/唯一约束使用 DROP CONSTRAINT，普通索引使用 DROP INDEX。
func (m *Manager) DropIndex(ctx context.Context, schemaName, tableName, indexName string) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	index, err := m.quote(indexName)
	if err != nil {
		return nil, err
	}
	var constraint bool
	if err = m.db.QueryRowContext(ctx, "SELECT CASE WHEN is_primary_key = 1 OR is_unique_constraint = 1 THEN 1 ELSE 0 END FROM sys.indexes WHERE object_id = OBJECT_ID(@p1) AND name = @p2", table, indexName).Scan(&constraint); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("当前表中不存在该索引")
	} else if err != nil {
		return nil, err
	}
	query := "DROP INDEX " + index + " ON " + table
	if constraint {
		query = "ALTER TABLE " + table + " DROP CONSTRAINT " + index
	}
	if _, err = m.db.ExecContext(ctx, query); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// DDL 读取视图的原始定义，用户表生成包含字段、主键、默认值和索引的结构脚本。
func (m *Manager) DDL(ctx context.Context, schemaName, tableName string, includeForeignKeys ...bool) (_ string, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", err
	}
	var kind string
	var definition sql.NullString
	if err = m.db.QueryRowContext(ctx, "SELECT o.type, OBJECT_DEFINITION(o.object_id) FROM sys.objects o JOIN sys.schemas s ON s.schema_id = o.schema_id WHERE s.name = @p1 AND o.name = @p2 AND o.is_ms_shipped = 0", schema, tableName).Scan(&kind, &definition); errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("数据表或视图不存在或当前账户无权访问")
	} else if err != nil {
		return "", err
	}
	if strings.TrimSpace(kind) == "V" {
		if !definition.Valid {
			return "", errors.New("视图定义已加密或当前账户无权查看")
		}
		return definition.String, nil
	}
	if strings.TrimSpace(kind) != "U" {
		return "", errors.New("只支持查看数据表和视图的结构定义")
	}
	data, err := m.Columns(ctx, schema, tableName)
	if err != nil {
		return "", err
	}
	definitions := make([]string, 0)
	for _, record := range data.([]map[string]interface{}) {
		name := fmt.Sprint(record["name"])
		quoted, _ := m.quote(name)
		if m.truth(record["generated_kind"]) {
			return "", errors.New("此表包含时态或账本系统生成字段，请使用原生数据库工具导出完整结构")
		}
		if m.truth(record["computed"]) {
			var expression string
			var persisted bool
			if err = m.db.QueryRowContext(ctx, "SELECT definition, is_persisted FROM sys.computed_columns WHERE object_id = OBJECT_ID(@p1) AND name = @p2", table, name).Scan(&expression, &persisted); err != nil {
				return "", err
			}
			value := quoted + " AS " + expression
			if persisted {
				value += " PERSISTED"
			}
			definitions = append(definitions, value)
			continue
		}
		column := quoted + " " + fmt.Sprint(record["type"])
		if record["collation"] != nil {
			collation, _ := m.quote(fmt.Sprint(record["collation"]))
			column += " COLLATE " + collation
		}
		if m.truth(record["auto_increment"]) {
			var seed, increment string
			if err = m.db.QueryRowContext(ctx, "SELECT CONVERT(VARCHAR(100), seed_value), CONVERT(VARCHAR(100), increment_value) FROM sys.identity_columns WHERE object_id = OBJECT_ID(@p1) AND name = @p2", table, name).Scan(&seed, &increment); err != nil {
				return "", err
			}
			column += " IDENTITY(" + seed + "," + increment + ")"
		}
		if m.truth(record["nullable"]) {
			column += " NULL"
		} else {
			column += " NOT NULL"
		}
		if record["default"] != nil {
			column += " DEFAULT " + fmt.Sprint(record["default"])
		}
		definitions = append(definitions, column)
	}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return "", err
	}
	filterDefinition := "i.filter_definition"
	if version < 10 {
		filterDefinition = "NULL AS filter_definition"
	}
	indexes, err := m.records(ctx, "SELECT i.name, i.is_unique, i.is_primary_key, i.is_unique_constraint, i.type_desc, "+filterDefinition+", ic.is_included_column, ic.is_descending_key, c.name AS column_name FROM sys.indexes i JOIN sys.index_columns ic ON ic.object_id = i.object_id AND ic.index_id = i.index_id JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id WHERE i.object_id = OBJECT_ID(@p1) AND i.name IS NOT NULL AND i.type IN (1, 2) ORDER BY i.index_id, ic.is_included_column, ic.key_ordinal, ic.index_column_id", table)
	if err != nil {
		return "", err
	}
	positions := make(map[string]int)
	groups := make([]map[string]interface{}, 0)
	for _, index := range indexes {
		name := fmt.Sprint(index["name"])
		position, found := positions[name]
		if !found {
			position = len(groups)
			positions[name] = position
			index["keys"], index["include"] = []string{}, []string{}
			groups = append(groups, index)
		}
		column, _ := m.quote(fmt.Sprint(index["column_name"]))
		field := "keys"
		if m.truth(index["is_included_column"]) {
			field = "include"
		} else if m.truth(index["is_descending_key"]) {
			column += " DESC"
		} else {
			column += " ASC"
		}
		groups[position][field] = append(groups[position][field].([]string), column)
	}
	var indexScript strings.Builder
	for _, index := range groups {
		name, _ := m.quote(fmt.Sprint(index["name"]))
		fields := "(" + strings.Join(index["keys"].([]string), ", ") + ")"
		if m.truth(index["is_primary_key"]) || m.truth(index["is_unique_constraint"]) {
			kind := "UNIQUE"
			if m.truth(index["is_primary_key"]) {
				kind = "PRIMARY KEY"
			}
			definitions = append(definitions, "CONSTRAINT "+name+" "+kind+" "+fmt.Sprint(index["type_desc"])+" "+fields)
			continue
		}
		query := "CREATE "
		if m.truth(index["is_unique"]) {
			query += "UNIQUE "
		}
		query += fmt.Sprint(index["type_desc"]) + " INDEX " + name + " ON " + table + " " + fields
		if include := index["include"].([]string); len(include) > 0 {
			query += " INCLUDE (" + strings.Join(include, ", ") + ")"
		}
		if index["filter_definition"] != nil {
			query += " WHERE " + fmt.Sprint(index["filter_definition"])
		}
		indexScript.WriteString(query + ";\n")
	}
	checks, err := m.records(ctx, "SELECT name, definition, is_disabled, is_not_trusted FROM sys.check_constraints WHERE parent_object_id = OBJECT_ID(@p1) ORDER BY name", table)
	if err != nil {
		return "", err
	}
	for _, check := range checks {
		name, _ := m.quote(fmt.Sprint(check["name"]))
		validation := "CHECK"
		if m.truth(check["is_not_trusted"]) || m.truth(check["is_disabled"]) {
			validation = "NOCHECK"
		}
		indexScript.WriteString("ALTER TABLE " + table + " WITH " + validation + " ADD CONSTRAINT " + name + " CHECK " + fmt.Sprint(check["definition"]) + ";\n")
		if m.truth(check["is_disabled"]) {
			indexScript.WriteString("ALTER TABLE " + table + " NOCHECK CONSTRAINT " + name + ";\n")
		}
	}
	if len(includeForeignKeys) == 0 || includeForeignKeys[0] {
		foreignKeys, err := m.Objects(ctx, schema, "foreign_key")
		if err != nil {
			return "", err
		}
		for _, record := range foreignKeys.([]map[string]interface{}) {
			if !strings.EqualFold(fmt.Sprint(record["table"]), tableName) {
				continue
			}
			key := ForeignKeyDefinition{
				Name:             fmt.Sprint(record["name"]),
				Columns:          record["columns"].([]string),
				ReferenceSchema:  fmt.Sprint(record["reference_schema"]),
				ReferenceTable:   fmt.Sprint(record["reference_table"]),
				ReferenceColumns: record["reference_columns"].([]string),
				OnDelete:         fmt.Sprint(record["on_delete"]),
				OnUpdate:         fmt.Sprint(record["on_update"]),
			}
			definition, err := m.foreignKeyDefinition(key)
			if err != nil {
				return "", err
			}
			validation := "CHECK"
			if m.truth(record["is_not_trusted"]) || m.truth(record["is_disabled"]) {
				validation = "NOCHECK"
			}
			indexScript.WriteString("ALTER TABLE " + table + " WITH " + validation + " ADD " + definition + ";\n")
			if m.truth(record["is_disabled"]) {
				name, _ := m.quote(fmt.Sprint(record["name"]))
				indexScript.WriteString("ALTER TABLE " + table + " NOCHECK CONSTRAINT " + name + ";\n")
			}
		}
	}
	script := "CREATE TABLE " + table + " (\n  " + strings.Join(definitions, ",\n  ") + "\n);\n\n" + indexScript.String()
	if len(script) > 1<<20 {
		return "", errors.New("结构定义超过 1 MB")
	}
	return strings.TrimSpace(script), nil
}

// CreateView 创建视图；语句只接受单条查询，保留 SqlServer 的批次要求。
func (m *Manager) CreateView(ctx context.Context, schemaName, viewName, statement string) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	schema, view, err := m.target(schemaName, viewName)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(viewName) > 128 || strings.EqualFold(schema, "sys") || strings.EqualFold(schema, "INFORMATION_SCHEMA") {
		return nil, errors.New("不能使用系统模式或超过 128 字符的视图名称")
	}
	statement = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(statement), ";"))
	if statement == "" || len(statement) > 1<<20 || strings.ContainsRune(statement, 0) {
		return nil, errors.New("视图内容必须是单条查询，且不超过 1 MB")
	}
	tokens, err := m.schemaTokens(statement)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 || (tokens[0] != "SELECT" && tokens[0] != "WITH") {
		return nil, errors.New("视图内容必须为 SELECT 查询")
	}
	for _, token := range tokens {
		switch token {
		case "INSERT", "UPDATE", "DELETE", "MERGE", "CREATE", "ALTER", "DROP", "EXEC", "EXECUTE", "USE", "SET", "GRANT", "REVOKE", "DENY", "TRUNCATE", "GO", ";":
			return nil, errors.New("视图定义不能包含附加语句")
		}
	}
	if _, err = m.db.ExecContext(ctx, "CREATE VIEW "+view+" AS "+statement); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}
