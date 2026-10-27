package mysql

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
var columnTypePattern = regexp.MustCompile(`(?i)^(?:(?:TINYINT|SMALLINT|MEDIUMINT|INT|INTEGER|BIGINT|DECIMAL|DEC|NUMERIC|FIXED|FLOAT|DOUBLE(?:\s+PRECISION)?|REAL|BIT)(?:\(\s*\d+(?:\s*,\s*\d+)?\s*\))?(?:\s+(?:UNSIGNED|ZEROFILL))*|(?:BOOL|BOOLEAN)|(?:DATE|YEAR|TIME|DATETIME|TIMESTAMP)(?:\(\s*\d+\s*\))?|(?:CHAR|CHARACTER|VARCHAR|NCHAR|NVARCHAR|BINARY|VARBINARY)(?:\(\s*\d+\s*\))?(?:\s+BINARY)?|(?:TINYBLOB|BLOB|MEDIUMBLOB|LONGBLOB|TINYTEXT|TEXT|MEDIUMTEXT|LONGTEXT|JSON|GEOMETRY|POINT|LINESTRING|POLYGON|MULTIPOINT|MULTILINESTRING|MULTIPOLYGON|GEOMETRYCOLLECTION)|(?:ENUM|SET)\(\s*'(?:''|[^'\\])*'(?:\s*,\s*'(?:''|[^'\\])*')*\s*\))$`)
var integerTypePattern = regexp.MustCompile(`(?i)^(?:TINYINT|SMALLINT|MEDIUMINT|INT|INTEGER|BIGINT)(?:\s|\(|$)`)
var defaultLiteralPattern = regexp.MustCompile(`(?i)^(?:CURRENT_TIMESTAMP(?:\(\d*\))?|NULL|TRUE|FALSE|[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?)$`)

// columnSQL 只生成一个字段定义，防止类型或默认表达式改变 DDL 的结构。
func (m *Manager) columnSQL(ctx context.Context, column ColumnDefinition) (string, error) {
	name, err := m.quote(column.Name)
	if err != nil || utf8.RuneCountInString(column.Name) > 64 {
		return "", errors.New("字段名称必须为 1 至 64 个字符")
	}
	column.Type = strings.TrimSpace(column.Type)
	if len(column.Type) > 4096 || !columnTypePattern.MatchString(column.Type) {
		return "", errors.New("字段类型不合法，请仅填写 MySQL 类型及长度，例如 varchar(255) 或 bigint unsigned")
	}
	definition := name + " " + column.Type
	if column.Nullable {
		definition += " NULL"
	} else {
		definition += " NOT NULL"
	}
	if column.AutoIncrement {
		if column.Nullable || column.Default != nil || !integerTypePattern.MatchString(column.Type) {
			return "", errors.New("自增字段必须为非空整数，且不能设置默认值")
		}
		definition += " AUTO_INCREMENT"
	}
	if column.Default != nil {
		if len(*column.Default) > 64<<10 || strings.ContainsRune(*column.Default, 0) {
			return "", errors.New("字段默认值不合法或超过 64 KB")
		}
		if column.DefaultExpression {
			expression := strings.TrimSpace(*column.Default)
			if err := m.schemaExpression(expression); err != nil {
				return "", err
			}
			// 时间函数和常量使用旧版也支持的语法；真正的表达式按服务器能力判断。
			if defaultLiteralPattern.MatchString(expression) {
				definition += " DEFAULT " + expression
			} else {
				version, maria, err := m.serverVersion(ctx)
				if err != nil {
					return "", err
				}
				if (!maria && version < 80013) || (maria && version < 100201) {
					return "", errors.New("此服务器版本不支持表达式默认值，请使用固定默认值或 CURRENT_TIMESTAMP")
				}
				definition += " DEFAULT (" + expression + ")"
			}
		} else {
			literal, err := m.schemaLiteral(ctx, *column.Default)
			if err != nil {
				return "", err
			}
			definition += " DEFAULT " + literal
		}
	}
	if column.Comment != nil {
		version, _, err := m.serverVersion(ctx)
		if err != nil {
			return "", err
		}
		limit := 1024
		if version < 50503 {
			limit = 255
		}
		if !utf8.ValidString(*column.Comment) || utf8.RuneCountInString(*column.Comment) > limit || strings.ContainsRune(*column.Comment, 0) {
			return "", fmt.Errorf("当前版本的字段备注不能超过 %d 个字符，且不能包含空字符", limit)
		}
		literal, err := m.schemaLiteral(ctx, *column.Comment)
		if err != nil {
			return "", err
		}
		definition += " COMMENT " + literal
	}
	return definition, nil
}

// schemaExpression 校验完整表达式的括号与引号，不拆分或执行额外语句。
func (m *Manager) schemaExpression(expression string) error {
	if expression == "" || len(expression) > 1<<20 || strings.ContainsRune(expression, 0) {
		return errors.New("SQL 表达式不能为空且不能超过 1 MB")
	}
	depth, quote := 0, byte(0)
	for index := 0; index < len(expression); index++ {
		ch := expression[index]
		if quote != 0 {
			if ch == '\\' {
				return errors.New("默认表达式中的反斜杠请改用标准 SQL 字面量")
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
		case '\'', '"', '`':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return errors.New("默认表达式括号不匹配")
			}
		case ';', '#':
			return errors.New("默认表达式不能包含多条语句或注释")
		case '-', '/':
			if index+1 < len(expression) && ((ch == '-' && expression[index+1] == '-') || (ch == '/' && expression[index+1] == '*')) {
				return errors.New("默认表达式不能包含注释")
			}
		}
	}
	if depth != 0 || quote != 0 {
		return errors.New("默认表达式的括号或引号不匹配")
	}
	return nil
}

// schemaLiteral 按当前 SQL 模式转义字符串，兼容 NO_BACKSLASH_ESCAPES。
func (m *Manager) schemaLiteral(ctx context.Context, value string) (string, error) {
	if strings.Contains(value, "\\") {
		var mode string
		if err := m.db.QueryRowContext(ctx, "SELECT @@SESSION.sql_mode").Scan(&mode); err != nil {
			return "", err
		}
		if !strings.Contains(strings.ToUpper(mode), "NO_BACKSLASH_ESCAPES") {
			value = strings.ReplaceAll(value, "\\", "\\\\")
		}
	}
	return "'" + strings.ReplaceAll(value, "'", "''") + "'", nil
}

// CreateTable 同时创建真实 MySQL 表及索引、外键，复合索引按提交顺序建立。
func (m *Manager) CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(schema) {
	case "", "mysql", "information_schema", "performance_schema", "sys":
		return nil, errors.New("不能在系统数据库中创建表")
	}
	if utf8.RuneCountInString(tableName) > 64 || len(columns) == 0 || len(columns) > 1024 {
		return nil, errors.New("表名不能超过 64 个字符，字段数量必须为 1 至 1024 个")
	}
	if len(primaryKeys) > 16 {
		return nil, errors.New("联合主键不能超过 16 个字段")
	}
	if len(indexes) > 64 {
		return nil, errors.New("索引数量不能超过 64 个")
	}
	definitions := make([]string, 0, len(columns)+len(indexes)+len(foreignKeys)+1)
	known := make(map[string]bool, len(columns))
	primary := make(map[string]bool, len(primaryKeys))
	for _, key := range primaryKeys {
		if primary[strings.ToLower(key)] {
			return nil, errors.New("主键字段不能重复")
		}
		primary[strings.ToLower(key)] = true
	}
	autoCount := 0
	for _, column := range columns {
		key := strings.ToLower(column.Name)
		if known[key] {
			return nil, errors.New("字段名称不能重复")
		}
		known[key] = true
		if primary[key] {
			column.Nullable = false
		}
		if column.AutoIncrement {
			autoCount++
			if len(primaryKeys) == 0 || !strings.EqualFold(primaryKeys[0], column.Name) {
				return nil, errors.New("自增字段必须是主键的第一个字段")
			}
		}
		definition, buildErr := m.columnSQL(ctx, column)
		if buildErr != nil {
			return nil, fmt.Errorf("字段 %s：%w", column.Name, buildErr)
		}
		definitions = append(definitions, definition)
	}
	if autoCount > 1 {
		return nil, errors.New("一个表只能有一个自增字段")
	}
	if len(primaryKeys) > 0 {
		names := make([]string, 0, len(primaryKeys))
		for _, key := range primaryKeys {
			if !known[strings.ToLower(key)] {
				return nil, errors.New("主键字段不存在")
			}
			name, _ := m.quote(key)
			names = append(names, name)
		}
		definitions = append(definitions, "PRIMARY KEY ("+strings.Join(names, ",")+")")
	}

	indexNames := make(map[string]bool, len(indexes))
	for _, index := range indexes {
		name, quoteErr := m.quote(index.Name)
		if quoteErr != nil || !utf8.ValidString(index.Name) || utf8.RuneCountInString(index.Name) > 64 {
			return nil, errors.New("索引名称必须为 1 至 64 个字符")
		}
		key := strings.ToLower(index.Name)
		if key == "primary" {
			return nil, errors.New("PRIMARY 是主键保留名称")
		}
		if indexNames[key] {
			return nil, errors.New("索引名称不能重复")
		}
		indexNames[key] = true
		if len(index.Columns) == 0 || len(index.Columns) > 16 {
			return nil, errors.New("索引字段数量必须为 1 至 16 个")
		}
		names := make([]string, 0, len(index.Columns))
		seen := make(map[string]bool, len(index.Columns))
		for _, column := range index.Columns {
			key := strings.ToLower(column.Name)
			if !known[key] {
				return nil, errors.New("索引字段不存在")
			}
			if seen[key] {
				return nil, errors.New("索引字段不能重复")
			}
			seen[key] = true
			name, _ := m.quote(column.Name)
			if column.Desc {
				name += " DESC"
			}
			names = append(names, name)
		}
		clause := "INDEX " + name + " (" + strings.Join(names, ",") + ")"
		if index.Unique {
			clause = "UNIQUE " + clause
		}
		if index.Comment != "" {
			if !utf8.ValidString(index.Comment) || utf8.RuneCountInString(index.Comment) > 1024 || strings.ContainsRune(index.Comment, 0) {
				return nil, errors.New("索引备注不能超过 1024 个字符，且不能包含空字符")
			}
			version, _, versionErr := m.serverVersion(ctx)
			if versionErr != nil {
				return nil, versionErr
			}
			if version < 50503 {
				return nil, errors.New("当前数据库版本不支持索引备注")
			}
			literal, literalErr := m.schemaLiteral(ctx, index.Comment)
			if literalErr != nil {
				return nil, literalErr
			}
			clause += " COMMENT " + literal
		}
		definitions = append(definitions, clause)
	}

	if len(foreignKeys) > 64 {
		return nil, errors.New("外键数量不能超过 64 个")
	}
	constraints := make(map[string]bool, len(foreignKeys))
	for _, foreignKey := range foreignKeys {
		name := foreignKey.Name
		if constraints[strings.ToLower(name)] {
			return nil, errors.New("外键名称不能重复")
		}
		constraints[strings.ToLower(name)] = true
		for _, name := range foreignKey.Columns {
			if !known[strings.ToLower(name)] {
				return nil, errors.New("外键字段不存在")
			}
		}
		clause, buildErr := m.foreignKeySQL(schema, foreignKey)
		if buildErr != nil {
			return nil, buildErr
		}
		definitions = append(definitions, clause)
	}
	engine := ""
	if len(foreignKeys) > 0 {
		// 老机器可能默认使用 MyISAM，指定有外键能力的引擎，防止约束被静默忽略。
		engines, err := m.records(ctx, "SHOW ENGINES")
		if err != nil {
			return nil, err
		}
		for _, row := range engines {
			if strings.EqualFold(fmt.Sprint(row["engine"]), "InnoDB") && (strings.EqualFold(fmt.Sprint(row["support"]), "YES") || strings.EqualFold(fmt.Sprint(row["support"]), "DEFAULT")) {
				engine = " ENGINE=InnoDB"
				break
			}
		}
		if engine == "" {
			return nil, errors.New("服务器未启用支持外键的 InnoDB 引擎")
		}
	}
	return m.executeMaintenance(ctx, "CREATE TABLE "+table+" ("+strings.Join(definitions, ", ")+")"+engine)
}

// RenameTable 重命名真实表或视图，数据库原生维护其索引与约束。
func (m *Manager) RenameTable(ctx context.Context, schemaName, tableName, newName string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, _, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	_, target, err := m.target(schemaName, newName)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(newName) > 64 {
		return nil, errors.New("表名称不能超过 64 个字符")
	}
	if tableName == newName {
		return nil, errors.New("新表名称与当前名称相同")
	}
	return m.executeMaintenance(ctx, "RENAME TABLE "+table+" TO "+target)
}

// AddColumn 新增真实字段，重复字段和版本限制交由 MySQL 原生验证。
func (m *Manager) AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "BASE TABLE" {
		return nil, errors.New("只能为数据表添加字段")
	}
	definition, err := m.columnSQL(ctx, column)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" ADD COLUMN "+definition)
}

// UpdateColumn 修改字段，保留未提交的备注、排序规则和 ON UPDATE 属性。
func (m *Manager) UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "BASE TABLE" {
		return nil, errors.New("只能修改数据表的字段")
	}
	name, err := m.quote(columnName)
	if err != nil {
		return nil, err
	}
	schema, _, _ := m.target(schemaName, tableName)
	var extra, comment string
	var characterSet, collation sql.NullString
	err = m.db.QueryRowContext(ctx, "SELECT EXTRA, COLUMN_COMMENT, CHARACTER_SET_NAME, COLLATION_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?", schema, tableName, columnName).Scan(&extra, &comment, &characterSet, &collation)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("字段不存在")
	}
	if err != nil {
		return nil, err
	}
	upper := strings.ToUpper(extra)
	if regexp.MustCompile(`(?:VIRTUAL|STORED|PERSISTENT) GENERATED`).MatchString(upper) {
		return nil, errors.New("生成列请使用 SQL 编辑器修改，避免丢失生成表达式")
	}
	if column.Comment == nil {
		column.Comment = &comment
	}
	definition, err := m.columnSQL(ctx, column)
	if err != nil {
		return nil, err
	}
	// 排序规则仅在新类型仍为字符类型时保留，转为数值或二进制时不携带字符属性。
	if characterSet.Valid && regexp.MustCompile(`(?i)^(?:CHAR|CHARACTER|VARCHAR|NCHAR|NVARCHAR|TINYTEXT|TEXT|MEDIUMTEXT|LONGTEXT|ENUM|SET)(?:\s|\(|$)`).MatchString(strings.TrimSpace(column.Type)) {
		charset, quoteErr := m.quote(characterSet.String)
		if quoteErr != nil {
			return nil, quoteErr
		}
		rule, quoteErr := m.quote(collation.String)
		if quoteErr != nil {
			return nil, quoteErr
		}
		// 字符属性属于类型，必须放在 NULL / DEFAULT 之前。
		quoted, _ := m.quote(column.Name)
		prefix := quoted + " " + strings.TrimSpace(column.Type)
		definition = prefix + " CHARACTER SET " + charset + " COLLATE " + rule + strings.TrimPrefix(definition, prefix)
	}
	if position := strings.Index(strings.ToLower(extra), "on update "); position >= 0 {
		update := strings.Fields(extra[position+len("on update "):])[0]
		if !regexp.MustCompile(`(?i)^CURRENT_TIMESTAMP(?:\(\d*\))?$`).MatchString(update) {
			return nil, errors.New("当前字段的 ON UPDATE 表达式请通过 SQL 编辑器修改")
		}
		if !regexp.MustCompile(`(?i)^(?:DATETIME|TIMESTAMP)(?:\s|\(|$)`).MatchString(column.Type) {
			return nil, errors.New("具有 ON UPDATE 属性的字段请先通过 SQL 移除该属性，再修改类型")
		}
		definition += " ON UPDATE " + update
	}
	if strings.Contains(upper, "INVISIBLE") {
		definition += " INVISIBLE"
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" CHANGE COLUMN "+name+" "+definition)
}

// DropColumn 删除真实字段，不自动级联删除其他表的约束。
func (m *Manager) DropColumn(ctx context.Context, schemaName, tableName, columnName string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "BASE TABLE" {
		return nil, errors.New("只能删除数据表的字段")
	}
	name, err := m.quote(columnName)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" DROP COLUMN "+name)
}

// CreateIndex 创建普通、唯一索引或主键。
func (m *Manager) CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool, comment string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "BASE TABLE" {
		return nil, errors.New("只能为数据表创建索引")
	}
	if len(columns) == 0 || len(columns) > 16 {
		return nil, errors.New("索引字段数量必须为 1 至 16 个")
	}
	names := make([]string, 0, len(columns))
	known := make(map[string]bool, len(columns))
	for _, column := range columns {
		key := strings.ToLower(column.Name)
		if known[key] {
			return nil, errors.New("索引字段不能重复")
		}
		known[key] = true
		name, quoteErr := m.quote(column.Name)
		if quoteErr != nil {
			return nil, quoteErr
		}
		if column.Desc {
			name += " DESC"
		}
		names = append(names, name)
	}
	clause := "PRIMARY KEY"
	if !primary {
		name, quoteErr := m.quote(indexName)
		if quoteErr != nil || utf8.RuneCountInString(indexName) > 64 {
			return nil, errors.New("索引名称必须为 1 至 64 个字符")
		}
		if strings.EqualFold(indexName, "PRIMARY") {
			return nil, errors.New("PRIMARY 是主键保留名称")
		}
		clause = "INDEX " + name
		if unique {
			clause = "UNIQUE " + clause
		}
	}
	clause += " (" + strings.Join(names, ",") + ")"
	algorithm := ""
	if comment != "" {
		version, maria, versionErr := m.serverVersion(ctx)
		if versionErr != nil {
			return nil, versionErr
		}
		if version < 50503 {
			return nil, errors.New("当前数据库版本不支持索引备注")
		}
		if !utf8.ValidString(comment) || utf8.RuneCountInString(comment) > 1024 || strings.ContainsRune(comment, 0) {
			return nil, errors.New("索引备注不能超过 1024 个字符，且不能包含空字符")
		}
		literal, literalErr := m.schemaLiteral(ctx, comment)
		if literalErr != nil {
			return nil, literalErr
		}
		clause += " COMMENT " + literal
		if !maria && version >= 50600 && version < 50706 {
			algorithm = ", ALGORITHM=COPY"
		}
	}
	out, err = m.executeMaintenance(ctx, "ALTER TABLE "+table+" ADD "+clause+algorithm)
	if err != nil || comment == "" {
		return out, err
	}
	schema, _, _ := m.target(schemaName, tableName)
	if primary {
		indexName = "PRIMARY"
	}
	var saved string
	if err = m.db.QueryRowContext(ctx, "SELECT INDEX_COMMENT FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND INDEX_NAME = ? LIMIT 1", schema, tableName, indexName).Scan(&saved); err != nil {
		return nil, err
	}
	if saved != comment {
		return nil, errors.New("索引已创建，但当前数据库未保存备注，请刷新后修改备注")
	}
	return out, nil
}

// UpdateIndex 修改索引备注，沿用服务器生成的完整定义，保留前缀、表达式、可见性等属性。
func (m *Manager) UpdateIndex(ctx context.Context, schemaName, tableName, indexName, comment string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "BASE TABLE" {
		return nil, errors.New("只能修改数据表的索引")
	}
	name, err := m.quote(indexName)
	if err != nil {
		return nil, err
	}
	version, maria, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	if version < 50503 {
		return nil, errors.New("当前数据库版本不支持索引备注")
	}
	if !utf8.ValidString(comment) || utf8.RuneCountInString(comment) > 1024 || strings.ContainsRune(comment, 0) {
		return nil, errors.New("索引备注不能超过 1024 个字符，且不能包含空字符")
	}
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	schema, _, _ := m.target(schemaName, tableName)
	var saved string
	if err = conn.QueryRowContext(ctx, "SELECT INDEX_COMMENT FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND INDEX_NAME = ? LIMIT 1", schema, tableName, indexName).Scan(&saved); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("索引不存在")
		}
		return nil, err
	}
	if saved == comment {
		return &Result{
			Columns: []string{},
			Rows:    [][]interface{}{},
		}, nil
	}
	var mode string
	var quoted bool
	if err = conn.QueryRowContext(ctx, "SELECT @@SESSION.sql_mode, @@SESSION.sql_quote_show_create").Scan(&mode, &quoted); err != nil {
		return nil, err
	}
	if !quoted {
		if _, err = conn.ExecContext(ctx, "SET SESSION sql_quote_show_create = 1"); err != nil {
			return nil, err
		}
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, _ = conn.ExecContext(cleanup, "SET SESSION sql_quote_show_create = 0")
		}()
	}
	// 兼容模式可能省略索引选项；此时不重建，避免丢失原有配置。
	if strings.Contains(","+strings.ToUpper(mode)+",", ",NO_KEY_OPTIONS,") || maria && regexp.MustCompile(`(?i)(?:^|,)(?:ANSI|DB2|MAXDB|MSSQL|ORACLE|POSTGRESQL)(?:,|$)`).MatchString(mode) {
		return nil, errors.New("当前 SQL 模式会省略索引定义，请通过 SQL 编辑器修改索引备注")
	}
	var object, ddl string
	if err = conn.QueryRowContext(ctx, "SHOW CREATE TABLE "+table).Scan(&object, &ddl); err != nil {
		return nil, err
	}
	if len(ddl) > 8<<20 {
		return nil, errors.New("对象定义超过 8 MB")
	}
	noBackslash := strings.Contains(strings.ToUpper(mode), "NO_BACKSLASH_ESCAPES")
	ansiQuotes := strings.Contains(strings.ToUpper(mode), "ANSI_QUOTES")
	// 标记原定义中的词和括号层级；字符串里的逗号、换行和 COMMENT 不参与拆分。
	tokens := func(value string) ([][3]int, error) {
		result := make([][3]int, 0)
		depth := 0
		for i := 0; i < len(value); {
			if strings.ContainsRune(" \t\r\n", rune(value[i])) {
				i++
				continue
			}
			start := i
			if value[i] == '/' && i+1 < len(value) && value[i+1] == '*' {
				end := strings.Index(value[i+2:], "*/")
				if end < 0 {
					return nil, errors.New("索引定义中的注释不完整")
				}
				i += end + 4
				continue
			}
			switch ch := value[i]; ch {
			case '\'', '"', '`':
				i++
				closed := false
				for i < len(value) {
					if value[i] == '\\' && ch != '`' && !(ch == '"' && ansiQuotes) && !noBackslash {
						i += 2
						continue
					}
					if value[i] == ch {
						i++
						if i < len(value) && value[i] == ch {
							i++
							continue
						}
						closed = true
						break
					}
					i++
				}
				if !closed {
					return nil, errors.New("索引定义中的引号不完整")
				}
			case '(':
				result = append(result, [3]int{i, i + 1, depth})
				depth++
				i++
				continue
			case ')':
				depth--
				i++
			case ',':
				i++
			default:
				for i < len(value) && !strings.ContainsRune(" \t\r\n(),\"'`", rune(value[i])) && !(value[i] == '/' && i+1 < len(value) && value[i+1] == '*') {
					i++
				}
			}
			if depth < 0 {
				return nil, errors.New("索引定义中的括号不完整")
			}
			result = append(result, [3]int{start, i, depth})
		}
		if depth != 0 {
			return nil, errors.New("索引定义中的括号不完整")
		}
		return result, nil
	}
	parts, err := tokens(ddl)
	if err != nil {
		return nil, err
	}
	start, definition := -1, ""
	for _, part := range parts {
		word := ddl[part[0]:part[1]]
		if word == "(" && part[2] == 0 && start < 0 {
			start = part[1]
			continue
		}
		if start < 0 || !(word == "," && part[2] == 1 || word == ")" && part[2] == 0) {
			continue
		}
		candidate := strings.TrimSpace(ddl[start:part[0]])
		items, scanErr := tokens(candidate)
		if scanErr != nil {
			return nil, scanErr
		}
		if len(items) >= 2 {
			first, second := strings.ToUpper(candidate[items[0][0]:items[0][1]]), strings.ToUpper(candidate[items[1][0]:items[1][1]])
			if first == "PRIMARY" && second == "KEY" && strings.EqualFold(indexName, "PRIMARY") {
				definition = candidate
			} else {
				position := 0
				if first == "UNIQUE" || first == "FULLTEXT" || first == "SPATIAL" {
					position++
				}
				keyword := strings.ToUpper(candidate[items[position][0]:items[position][1]])
				if (keyword == "KEY" || keyword == "INDEX") && position+1 < len(items) {
					key := candidate[items[position+1][0]:items[position+1][1]]
					if len(key) >= 2 && (key[0] == '`' || key[0] == '"') {
						quote := key[:1]
						key = strings.ReplaceAll(key[1:len(key)-1], quote+quote, quote)
					}
					if strings.EqualFold(key, indexName) {
						definition = candidate
					}
				}
			}
		}
		if definition != "" || word == ")" {
			break
		}
		start = part[1]
	}
	if definition == "" {
		return nil, errors.New("未找到完整索引定义，请刷新后重试")
	}
	parts, err = tokens(definition)
	if err != nil {
		return nil, err
	}
	value := comment
	if !noBackslash {
		value = strings.ReplaceAll(value, "\\", "\\\\")
	}
	literal := "'" + strings.ReplaceAll(value, "'", "''") + "'"
	replaced := false
	for i, part := range parts {
		if part[2] == 0 && strings.EqualFold(definition[part[0]:part[1]], "COMMENT") {
			if i+1 >= len(parts) || definition[parts[i+1][0]] != '\'' {
				return nil, errors.New("无法识别原索引备注，请通过 SQL 编辑器修改")
			}
			definition = definition[:part[0]] + "COMMENT " + literal + definition[parts[i+1][1]:]
			replaced = true
			break
		}
	}
	if !replaced {
		definition += " COMMENT " + literal
	}
	drop := "INDEX " + name
	if strings.EqualFold(indexName, "PRIMARY") {
		drop = "PRIMARY KEY"
	}
	statement := "ALTER TABLE " + table + " DROP " + drop + ", ADD " + definition
	if !maria && version >= 50600 && version < 50706 {
		// 旧版 InnoDB 原地修改可能忽略备注，改用复制算法并在执行后核对。
		statement += ", ALGORITHM=COPY"
	}
	started := time.Now()
	result, err := conn.ExecContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	if err = conn.QueryRowContext(ctx, "SELECT INDEX_COMMENT FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND INDEX_NAME = ? LIMIT 1", schema, tableName, indexName).Scan(&saved); err != nil {
		return nil, err
	}
	if saved != comment {
		return nil, errors.New("当前数据库未保存索引备注，请通过 SQL 编辑器修改")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		Affected:  affected,
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// DropIndex 删除指定索引；主键使用 MySQL 的 DROP PRIMARY KEY。
func (m *Manager) DropIndex(ctx context.Context, schemaName, tableName, indexName string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "BASE TABLE" {
		return nil, errors.New("只能删除数据表的索引")
	}
	clause := "PRIMARY KEY"
	if !strings.EqualFold(indexName, "PRIMARY") {
		name, quoteErr := m.quote(indexName)
		if quoteErr != nil {
			return nil, quoteErr
		}
		clause = "INDEX " + name
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" DROP "+clause)
}

// DDL 返回服务器生成的建表或建视图语句。
func (m *Manager) DDL(ctx context.Context, schemaName, tableName string) (out string, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", err
	}
	rows, err := m.db.QueryContext(ctx, "SHOW CREATE TABLE "+table)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	names, err := rows.Columns()
	if err != nil {
		return "", err
	}
	if len(names) < 2 {
		return "", errors.New("服务器未返回对象定义")
	}
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return "", err
		}
		return "", errors.New("表或视图不存在")
	}
	values := make([]interface{}, len(names))
	pointers := make([]interface{}, len(names))
	for index := range values {
		pointers[index] = &values[index]
	}
	if err = rows.Scan(pointers...); err != nil {
		return "", err
	}
	switch value := values[1].(type) {
	case []byte:
		out = string(value)
	case string:
		out = value
	default:
		return "", errors.New("服务器返回的对象定义格式不正确")
	}
	if len(out) > 8<<20 {
		return "", errors.New("对象定义超过 8 MB")
	}
	return out + ";", nil
}
