package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 字段批量建表时复用校验表达式，避免为每个字段重新编译。
var columnTypePattern = regexp.MustCompile(`(?i)^(?:[a-z_][a-z0-9_$]*|"(?:[^"]|"")+")(?:(?:\s*\.\s*|\s+)(?:[a-z_][a-z0-9_$]*|"(?:[^"]|"")+"))*(?:\s*\(\s*\d+(?:\s*,\s*-?\d+)?\s*\))?(?:\s+(?:with|without)\s+time\s+zone)?(?:\s*\[\s*\d*\s*\])*$`)

// columnSQL 生成一个字段定义，类型名称支持自定义类型、精度和数组。
func (m *Manager) columnSQL(column ColumnDefinition) (string, error) {
	name, err := m.quote(column.Name)
	if err != nil || len(column.Name) > 63 {
		return "", errors.New("字段名称必须为 1 至 63 字节")
	}
	column.Type = strings.TrimSpace(column.Type)
	if len(column.Type) > 1024 || !columnTypePattern.MatchString(column.Type) {
		return "", errors.New("字段类型不合法，请仅填写 Postgres 类型，例如 varchar(255) 或 numeric(18,2)")
	}
	for _, word := range strings.Fields(strings.ToUpper(column.Type)) {
		switch word {
		case "DEFAULT", "NULL", "NOT", "PRIMARY", "UNIQUE", "CHECK", "REFERENCES", "GENERATED", "CONSTRAINT", "COLLATE":
			return "", errors.New("字段类型中不能包含约束或默认值")
		}
	}
	fieldType := column.Type
	if column.AutoIncrement {
		if column.Nullable || column.Default != nil {
			return "", errors.New("自增字段必须非空，且不能同时设置默认值")
		}
		switch strings.ToLower(column.Type) {
		case "smallint", "int2", "smallserial", "serial2":
			if m.version != 0 && m.version < 90200 {
				return "", errors.New("smallint 自增需要 Postgres 9.2 或以上版本，旧版本请使用 integer 自增")
			}
			fieldType = "smallserial"
		case "int", "integer", "int4", "serial", "serial4":
			fieldType = "serial"
		case "bigint", "int8", "bigserial", "serial8":
			fieldType = "bigserial"
		default:
			return "", errors.New("自增字段只支持 smallint、integer 或 bigint")
		}
	} else {
		switch strings.ToLower(column.Type) {
		case "serial", "serial2", "serial4", "serial8", "smallserial", "bigserial":
			return "", errors.New("serial 类型必须开启自增")
		}
	}
	definition := name + " " + fieldType
	if column.Nullable {
		definition += " NULL"
	} else {
		definition += " NOT NULL"
	}
	if column.Default != nil {
		if len(*column.Default) > 64<<10 || strings.ContainsRune(*column.Default, 0) {
			return "", errors.New("默认值不合法或超过 64 KB")
		}
		if column.DefaultExpression {
			expression := strings.TrimSpace(*column.Default)
			if err = m.schemaExpression(expression); err != nil {
				return "", err
			}
			definition += " DEFAULT (" + expression + ")"
		} else {
			definition += " DEFAULT E'" + strings.ReplaceAll(strings.ReplaceAll(*column.Default, "\\", "\\\\"), "'", "''") + "'"
		}
	}
	return definition, nil
}

// schemaExpression 确保表达式引号、括号完整，并拒绝额外语句。
func (m *Manager) schemaExpression(expression string) error {
	if expression == "" || len(expression) > 1<<20 || strings.ContainsRune(expression, 0) {
		return errors.New("SQL 表达式不能为空且不能超过 1 MB")
	}
	depth, quote := 0, byte(0)
	for index := 0; index < len(expression); index++ {
		ch := expression[index]
		if quote != 0 {
			if ch == '\\' {
				return errors.New("表达式中的反斜杠请改用标准 SQL 字面量")
			}
			if ch == quote {
				if index+1 < len(expression) && expression[index+1] == quote {
					index++
				} else {
					quote = 0
				}
			}
			continue
		}
		switch ch {
		case '\'', '"':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return errors.New("表达式括号不匹配")
			}
		case ';':
			return errors.New("表达式不能包含多条语句")
		case '-', '/':
			if index+1 < len(expression) && ((ch == '-' && expression[index+1] == '-') || (ch == '/' && expression[index+1] == '*')) {
				return errors.New("表达式不能包含注释")
			}
		}
	}
	if depth != 0 || quote != 0 {
		return errors.New("表达式的括号或引号不匹配")
	}
	return nil
}

// CreateTable 创建真实 Postgres 表，自增使用兼容旧版本的 serial 类型。
func (m *Manager) CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") || m.database == "template0" || m.database == "template1" {
		return nil, errors.New("不能在系统模式中创建表")
	}
	if len(tableName) > 63 || len(columns) == 0 || len(columns) > 1600 {
		return nil, errors.New("表名称不能超过 63 字节，字段数量必须为 1 至 1600 个")
	}
	definitions := make([]string, 0, len(columns)+1)
	known := make(map[string]bool, len(columns))
	primary := make(map[string]bool, len(primaryKeys))
	for _, key := range primaryKeys {
		if primary[key] {
			return nil, errors.New("主键字段不能重复")
		}
		primary[key] = true
	}
	for _, column := range columns {
		if known[column.Name] {
			return nil, errors.New("字段名称不能重复")
		}
		known[column.Name] = true
		if primary[column.Name] {
			column.Nullable = false
		}
		definition, buildErr := m.columnSQL(column)
		if buildErr != nil {
			return nil, fmt.Errorf("字段 %s：%w", column.Name, buildErr)
		}
		definitions = append(definitions, definition)
	}
	if len(primaryKeys) > 0 {
		names := make([]string, 0, len(primaryKeys))
		for _, key := range primaryKeys {
			if !known[key] {
				return nil, errors.New("主键字段不存在")
			}
			name, _ := m.quote(key)
			names = append(names, name)
		}
		definitions = append(definitions, "PRIMARY KEY ("+strings.Join(names, ",")+")")
	}

	if len(foreignKeys) > 64 {
		return nil, errors.New("外键数量不能超过 64 个")
	}
	constraints := make(map[string]bool, len(foreignKeys))
	for _, foreignKey := range foreignKeys {
		name := foreignKey.Name
		if constraints[name] {
			return nil, errors.New("外键名称不能重复")
		}
		constraints[name] = true
		for _, name := range foreignKey.Columns {
			if !known[name] {
				return nil, errors.New("外键字段不存在")
			}
		}
		clause, buildErr := m.foreignKeySQL(schema, foreignKey)
		if buildErr != nil {
			return nil, buildErr
		}
		definitions = append(definitions, clause)
	}
	if len(indexes) > 64 {
		return nil, errors.New("索引数量不能超过 64 个")
	}
	statements := []string{"CREATE TABLE " + table + " (" + strings.Join(definitions, ", ") + ")"}
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
		Columns: []string{},
		Rows:    [][]interface{}{},
	}, nil
}

// RenameTable 重命名表、视图或物化视图。
func (m *Manager) RenameTable(ctx context.Context, schemaName, tableName, newName string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	name, err := m.quote(newName)
	if err != nil || len(newName) > 63 {
		return nil, errors.New("表名称必须为 1 至 63 字节")
	}
	if tableName == newName {
		return nil, errors.New("新表名称与当前名称相同")
	}
	clause := "TABLE"
	if kind == "v" {
		clause = "VIEW"
	} else if kind == "m" {
		clause = "MATERIALIZED VIEW"
	}
	return m.executeMaintenance(ctx, "ALTER "+clause+" "+table+" RENAME TO "+name)
}

// AddColumn 新增字段，兼容 serial 自增字段与自定义类型。
func (m *Manager) AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "r" && kind != "p" {
		return nil, errors.New("只能为数据表添加字段")
	}
	definition, err := m.columnSQL(column)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" ADD COLUMN "+definition)
}

// UpdateColumn 在同一事务中修改字段类型、默认值、非空属性及名称，失败整体回滚。
func (m *Manager) UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "r" && kind != "p" {
		return nil, errors.New("只能修改数据表的字段")
	}
	// 复用字段定义校验；已有自增列有实际序列默认值，因此无需重复创建 serial。
	validation := column
	if column.AutoIncrement {
		validation.Default = nil
		validation.DefaultExpression = false
	}
	if _, err = m.columnSQL(validation); err != nil {
		return nil, err
	}
	schema, _, _ := m.target(schemaName, tableName)
	data, err := m.Columns(ctx, schema, tableName)
	if err != nil {
		return nil, err
	}
	var current map[string]interface{}
	for _, value := range data.([]map[string]interface{}) {
		if value["name"] == columnName {
			current = value
			break
		}
	}
	if current == nil {
		return nil, errors.New("字段不存在")
	}
	identity := fmt.Sprint(current["identity"])
	if identity == "<nil>" {
		identity = ""
	}
	if m.truth(current["generated"]) && identity == "" {
		return nil, errors.New("生成列请使用 SQL 编辑器修改，避免丢失生成表达式")
	}
	auto := m.truth(current["auto_increment"])
	if column.AutoIncrement && !auto && column.Default != nil {
		return nil, errors.New("开启自增时不能同时设置默认值")
	}
	if auto && column.AutoIncrement && column.Nullable {
		return nil, errors.New("自增字段不能允许为空")
	}
	oldName, err := m.quote(columnName)
	if err != nil {
		return nil, err
	}
	newName, _ := m.quote(column.Name)
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	statements := make([]string, 0, 6)
	// 先移除旧默认值，再修改类型，可避免旧类型默认值阻止合法类型转换。
	if identity != "" && !column.AutoIncrement {
		statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" DROP IDENTITY")
	}
	if identity == "" {
		statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" DROP DEFAULT")
	}
	if !strings.EqualFold(strings.TrimSpace(column.Type), fmt.Sprint(current["type"])) {
		statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" TYPE "+strings.TrimSpace(column.Type)+" USING "+oldName+"::"+strings.TrimSpace(column.Type))
	}
	if column.AutoIncrement && !auto {
		var version string
		if err = tx.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version); err != nil {
			return nil, err
		}
		number, _ := strconv.Atoi(version)
		statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" SET NOT NULL")
		if number >= 100000 {
			statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" ADD GENERATED BY DEFAULT AS IDENTITY")
		} else {
			sequenceName := tableName + "_" + columnName + "_seq"
			if len(sequenceName) > 63 {
				return nil, errors.New("旧版本 Postgres 的序列名称超过 63 字节，请缩短表或字段名称")
			}
			sequence, quoteErr := m.quote(sequenceName)
			if quoteErr != nil {
				return nil, quoteErr
			}
			quotedSchema, _ := m.quote(schema)
			targetSequence := quotedSchema + "." + sequence
			statements = append(statements, "CREATE SEQUENCE "+targetSequence+" OWNED BY "+table+"."+oldName,
				"ALTER TABLE "+table+" ALTER COLUMN "+oldName+" SET DEFAULT nextval(E'"+strings.ReplaceAll(strings.ReplaceAll(targetSequence, "\\", "\\\\"), "'", "''")+"'::regclass)")
		}
	}
	if (identity == "" || !column.AutoIncrement) && !(column.AutoIncrement && !auto) {
		defaultSQL := ""
		if auto && column.AutoIncrement {
			if column.Default != nil && (!column.DefaultExpression || *column.Default != fmt.Sprint(current["default"])) {
				return nil, errors.New("自增字段默认值由原序列管理，不能手动修改")
			}
			defaultSQL = fmt.Sprint(current["default"])
		} else if column.Default != nil {
			if column.DefaultExpression {
				defaultSQL = strings.TrimSpace(*column.Default)
				if err = m.schemaExpression(defaultSQL); err != nil {
					return nil, err
				}
				defaultSQL = "(" + defaultSQL + ")"
			} else {
				defaultSQL = "E'" + strings.ReplaceAll(strings.ReplaceAll(*column.Default, "\\", "\\\\"), "'", "''") + "'"
			}
		}
		if defaultSQL != "" {
			statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" SET DEFAULT "+defaultSQL)
		}
	} else if column.Default != nil {
		return nil, errors.New("IDENTITY 字段不能设置默认值")
	}
	nullable := "SET NOT NULL"
	if column.Nullable {
		nullable = "DROP NOT NULL"
	}
	statements = append(statements, "ALTER TABLE "+table+" ALTER COLUMN "+oldName+" "+nullable)
	if columnName != column.Name {
		statements = append(statements, "ALTER TABLE "+table+" RENAME COLUMN "+oldName+" TO "+newName)
	}
	for _, statement := range statements {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return nil, err
		}
	}
	if column.AutoIncrement && !auto {
		// 新序列从现有最大值之后开始，避免首次自增与历史数据冲突。
		field, _ := m.quote(column.Name)
		var next int64
		if err = tx.QueryRowContext(ctx, "SELECT setval(pg_get_serial_sequence($1,$2), GREATEST(COALESCE((SELECT MAX("+field+") FROM "+table+"),0)+1,1), false)", table, column.Name).Scan(&next); err != nil {
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

// DropColumn 删除字段，不自动 CASCADE 删除依赖对象。
func (m *Manager) DropColumn(ctx context.Context, schemaName, tableName, columnName string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "r" && kind != "p" {
		return nil, errors.New("只能删除数据表的字段")
	}
	name, err := m.quote(columnName)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" DROP COLUMN "+name)
}

// CreateIndex 创建普通或唯一索引；primary 以真实主键约束建立。
func (m *Manager) CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "v" {
		return nil, errors.New("普通视图不能创建索引")
	}
	if len(columns) == 0 || len(columns) > 32 {
		return nil, errors.New("索引字段数量必须为 1 至 32 个")
	}
	name, err := m.quote(indexName)
	if err != nil || len(indexName) > 63 {
		return nil, errors.New("索引名称必须为 1 至 63 字节")
	}
	names := make([]string, 0, len(columns))
	known := make(map[string]bool, len(columns))
	for _, column := range columns {
		if known[column.Name] {
			return nil, errors.New("索引字段不能重复")
		}
		known[column.Name] = true
		field, quoteErr := m.quote(column.Name)
		if quoteErr != nil {
			return nil, quoteErr
		}
		if column.Desc {
			if primary {
				return nil, errors.New("主键不支持指定降序")
			}
			field += " DESC"
		}
		names = append(names, field)
	}
	if primary {
		if kind == "m" {
			return nil, errors.New("物化视图不能创建主键")
		}
		return m.executeMaintenance(ctx, "ALTER TABLE "+table+" ADD CONSTRAINT "+name+" PRIMARY KEY ("+strings.Join(names, ",")+")")
	}
	clause := "CREATE INDEX "
	if unique {
		clause = "CREATE UNIQUE INDEX "
	}
	return m.executeMaintenance(ctx, clause+name+" ON "+table+" ("+strings.Join(names, ",")+")")
}

// DropIndex 删除指定表上的索引；约束索引通过 DROP CONSTRAINT 删除，依赖限制由数据库处理。
func (m *Manager) DropIndex(ctx context.Context, schemaName, tableName, indexName string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "v" {
		return nil, errors.New("普通视图没有索引")
	}
	schema, _, _ := m.target(schemaName, tableName)
	var constraint sql.NullString
	err = m.db.QueryRowContext(ctx, "SELECT p.conname FROM pg_index x JOIN pg_class t ON t.oid = x.indrelid JOIN pg_namespace n ON n.oid = t.relnamespace JOIN pg_class i ON i.oid = x.indexrelid LEFT JOIN pg_depend dep ON dep.classid = 'pg_class'::regclass AND dep.objid = x.indexrelid AND dep.refclassid = 'pg_constraint'::regclass AND dep.deptype = 'i' LEFT JOIN pg_constraint p ON p.oid = dep.refobjid AND p.conrelid = t.oid AND p.contype IN ('p','u','x') WHERE n.nspname = $1 AND t.relname = $2 AND i.relname = $3", schema, tableName, indexName).Scan(&constraint)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("当前表不存在此索引")
	}
	if err != nil {
		return nil, err
	}
	if constraint.Valid {
		name, quoteErr := m.quote(constraint.String)
		if quoteErr != nil {
			return nil, quoteErr
		}
		return m.executeMaintenance(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+name)
	}
	quotedSchema, _ := m.quote(schema)
	name, err := m.quote(indexName)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "DROP INDEX "+quotedSchema+"."+name)
}

// DDL 从 Postgres 系统目录生成对象结构、约束和索引定义，不读取表数据。
func (m *Manager) DDL(ctx context.Context, schemaName, tableName string, exportConnection ...*sql.Conn) (out string, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	queryRow, records := m.db.QueryRowContext, m.records
	// 导出连接有固定 search_path 与快照；结构查询保持同一会话，并后置外键、索引和触发器。
	if len(exportConnection) > 0 && exportConnection[0] != nil {
		conn := exportConnection[0]
		queryRow = conn.QueryRowContext
		records = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			rows, err := conn.QueryContext(ctx, query, args...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			names, err := rows.Columns()
			if err != nil {
				return nil, err
			}
			values, pointers := make([]interface{}, len(names)), make([]interface{}, len(names))
			for i := range values {
				pointers[i] = &values[i]
			}
			result := []map[string]interface{}{}
			for rows.Next() {
				if len(result) >= 10000 {
					return nil, errors.New("结构条目超过 10000")
				}
				if err = rows.Scan(pointers...); err != nil {
					return nil, err
				}
				item := map[string]interface{}{}
				for i, value := range values {
					if data, ok := value.([]byte); ok {
						value = string(data)
					}
					item[names[i]] = value
				}
				result = append(result, item)
			}
			return result, rows.Err()
		}
	}
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", err
	}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return "", err
	}
	persistenceField := "c.relpersistence"
	if version < 90100 {
		persistenceField = "CASE WHEN c.relistemp THEN 't' ELSE 'p' END"
	}
	var oid int64
	var kind, persistence string
	err = queryRow(ctx, "SELECT c.oid, c.relkind, "+persistenceField+" FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2", schema, tableName).Scan(&oid, &kind, &persistence)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("对象不存在或无权访问")
	}
	if err != nil {
		return "", err
	}
	if kind == "v" || kind == "m" {
		var definition string
		if err = queryRow(ctx, "SELECT pg_get_viewdef($1::oid, true)", oid).Scan(&definition); err != nil {
			return "", err
		}
		clause := "CREATE VIEW "
		if kind == "m" {
			clause = "CREATE MATERIALIZED VIEW "
		}
		return clause + table + " AS\n" + strings.TrimSuffix(strings.TrimSpace(definition), ";") + ";", nil
	}
	if kind != "r" && kind != "p" {
		return "", errors.New("此对象类型不支持生成建表语句")
	}
	identityField, generatedField, collationField, collationJoins := "''", "''", "NULL::text", ""
	if version >= 100000 {
		identityField = "a.attidentity"
	}
	if version >= 120000 {
		generatedField = "a.attgenerated"
	}
	if version >= 90100 {
		collationField = "CASE WHEN a.attcollation <> t.typcollation THEN quote_ident(cn.nspname)||'.'||quote_ident(co.collname) END"
		collationJoins = " LEFT JOIN pg_collation co ON co.oid = a.attcollation LEFT JOIN pg_namespace cn ON cn.oid = co.collnamespace"
	}
	columns, err := records(ctx, "SELECT a.attname AS name, format_type(a.atttypid,a.atttypmod) AS type, a.attnotnull AS not_null, pg_get_expr(d.adbin,d.adrelid) AS default_value, EXISTS(SELECT 1 FROM pg_depend fd WHERE fd.classid='pg_attrdef'::regclass AND fd.objid=d.oid AND fd.refclassid='pg_proc'::regclass) AS function_default, "+identityField+" AS identity, "+generatedField+" AS generated, col_description(a.attrelid,a.attnum) AS comment, "+collationField+" AS collation FROM pg_attribute a JOIN pg_type t ON t.oid = a.atttypid LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum"+collationJoins+" WHERE a.attrelid = $1::oid AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum", oid)
	if err != nil {
		return "", err
	}
	definitions := make([]string, 0, len(columns))
	comments := make([]string, 0)
	for _, column := range columns {
		name, _ := m.quote(fmt.Sprint(column["name"]))
		definition := name + " " + fmt.Sprint(column["type"])
		if column["collation"] != nil {
			definition += " COLLATE " + fmt.Sprint(column["collation"])
		}
		identity := fmt.Sprint(column["identity"])
		generated := fmt.Sprint(column["generated"])
		if generated != "" && generated != "<nil>" {
			storage := "STORED"
			if generated == "v" {
				storage = "VIRTUAL"
			}
			definition += " GENERATED ALWAYS AS (" + fmt.Sprint(column["default_value"]) + ") " + storage
		} else if identity == "a" {
			definition += " GENERATED ALWAYS AS IDENTITY"
		} else if identity == "b" {
			definition += " GENERATED BY DEFAULT AS IDENTITY"
		} else if column["default_value"] != nil && (len(exportConnection) == 0 || exportConnection[0] == nil || !m.truth(column["function_default"])) {
			definition += " DEFAULT " + fmt.Sprint(column["default_value"])
		}
		if (identity == "a" || identity == "b") && len(exportConnection) > 0 && exportConnection[0] != nil {
			var sequence sql.NullString
			if err = queryRow(ctx, "SELECT pg_get_serial_sequence($1,$2)", table, fmt.Sprint(column["name"])).Scan(&sequence); err != nil {
				return "", err
			}
			if sequence.Valid {
				definition += " (SEQUENCE NAME " + sequence.String + ")"
			}
		}
		if m.truth(column["not_null"]) {
			definition += " NOT NULL"
		}
		definitions = append(definitions, "  "+definition)
		if column["comment"] != nil {
			comments = append(comments, "COMMENT ON COLUMN "+table+"."+name+" IS E'"+strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(column["comment"]), "\\", "\\\\"), "'", "''")+"';")
		}
	}
	constraintWhere := ""
	if len(exportConnection) > 0 && exportConnection[0] != nil {
		constraintWhere = " AND contype NOT IN ('f','c')"
	}
	constraints, err := records(ctx, "SELECT conname AS name, pg_get_constraintdef(oid, true) AS definition FROM pg_constraint WHERE conrelid = $1::oid"+constraintWhere+" ORDER BY conname", oid)
	if err != nil {
		return "", err
	}
	for _, constraint := range constraints {
		name, _ := m.quote(fmt.Sprint(constraint["name"]))
		definitions = append(definitions, "  CONSTRAINT "+name+" "+fmt.Sprint(constraint["definition"]))
	}
	clause := "CREATE TABLE "
	if persistence == "u" {
		clause = "CREATE UNLOGGED TABLE "
	} else if persistence == "t" {
		clause = "CREATE TEMPORARY TABLE "
	}
	out = "-- Postgres 对象结构；引用的类型与序列需要已存在。\n" + clause + table + " (\n" + strings.Join(definitions, ",\n") + "\n)"
	if kind == "p" {
		var partition string
		if err = queryRow(ctx, "SELECT pg_get_partkeydef($1::oid)", oid).Scan(&partition); err != nil {
			return "", err
		}
		out += " PARTITION BY " + partition
	}
	parents, err := records(ctx, "SELECT quote_ident(n.nspname)||'.'||quote_ident(c.relname) AS name FROM pg_inherits i JOIN pg_class c ON c.oid = i.inhparent JOIN pg_namespace n ON n.oid = c.relnamespace WHERE i.inhrelid = $1::oid ORDER BY i.inhseqno", oid)
	if err != nil {
		return "", err
	}
	if len(parents) > 0 {
		var version string
		if err = queryRow(ctx, "SHOW server_version_num").Scan(&version); err != nil {
			return "", err
		}
		number, _ := strconv.Atoi(version)
		partition := ""
		if number >= 100000 {
			var bound sql.NullString
			if err = queryRow(ctx, "SELECT pg_get_expr(relpartbound,oid) FROM pg_class WHERE oid = $1::oid", oid).Scan(&bound); err != nil {
				return "", err
			}
			partition = bound.String
		}
		if partition != "" {
			out = "-- Postgres 分区结构；父表需要已存在。\n" + clause + table + " PARTITION OF " + fmt.Sprint(parents[0]["name"]) + " " + partition
		} else {
			names := make([]string, 0, len(parents))
			for _, parent := range parents {
				names = append(names, fmt.Sprint(parent["name"]))
			}
			out += " INHERITS (" + strings.Join(names, ",") + ")"
		}
	}
	var options sql.NullString
	var tableComment sql.NullString
	if err = queryRow(ctx, "SELECT array_to_string(reloptions, ', '), obj_description(oid, 'pg_class') FROM pg_class WHERE oid = $1::oid", oid).Scan(&options, &tableComment); err != nil {
		return "", err
	}
	if options.Valid && options.String != "" {
		out += " WITH (" + options.String + ")"
	}
	out += ";"
	if len(exportConnection) == 0 || exportConnection[0] == nil {
		indexes, err := records(ctx, "SELECT pg_get_indexdef(x.indexrelid) AS definition FROM pg_index x WHERE x.indrelid = $1::oid AND NOT EXISTS (SELECT 1 FROM pg_constraint p JOIN pg_depend dep ON dep.refclassid = 'pg_constraint'::regclass AND dep.refobjid = p.oid AND dep.classid = 'pg_class'::regclass AND dep.deptype = 'i' WHERE dep.objid = x.indexrelid AND p.contype IN ('p','u','x')) ORDER BY x.indexrelid", oid)
		if err != nil {
			return "", err
		}
		for _, index := range indexes {
			out += "\n" + fmt.Sprint(index["definition"]) + ";"
		}
		triggerDefinition, triggerVisible := "pg_get_triggerdef(oid, true)", "NOT tgisinternal"
		if version < 90000 {
			triggerDefinition, triggerVisible = "pg_get_triggerdef(oid)", "tgconstraint = 0"
		}
		triggers, err := records(ctx, "SELECT "+triggerDefinition+" AS definition FROM pg_trigger WHERE tgrelid = $1::oid AND "+triggerVisible+" ORDER BY tgname", oid)
		if err != nil {
			return "", err
		}
		for _, trigger := range triggers {
			out += "\n" + fmt.Sprint(trigger["definition"]) + ";"
		}
	}
	if tableComment.Valid {
		comments = append(comments, "COMMENT ON TABLE "+table+" IS E'"+strings.ReplaceAll(strings.ReplaceAll(tableComment.String, "\\", "\\\\"), "'", "''")+"';")
	}
	if len(comments) > 0 {
		out += "\n" + strings.Join(comments, "\n")
	}
	if len(out) > 8<<20 {
		return "", errors.New("对象定义超过 8 MB")
	}
	return out, nil
}

// CreateView 创建真实 Postgres 视图，只接受一条 SELECT 或 WITH 查询。
func (m *Manager) CreateView(ctx context.Context, schemaName, viewName, statement string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, view, err := m.target(schemaName, viewName)
	if err != nil {
		return nil, err
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") || m.database == "template0" || m.database == "template1" {
		return nil, errors.New("不能在系统模式中创建视图")
	}
	statement = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(statement), ";"))
	if len(viewName) > 63 || !regexp.MustCompile(`(?i)^(?:SELECT|WITH)\b`).MatchString(statement) {
		return nil, errors.New("请输入 SELECT 或 WITH 查询，视图名称不能超过 63 字节")
	}
	if err = m.schemaExpression(statement); err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "CREATE VIEW "+view+" AS "+statement)
}
