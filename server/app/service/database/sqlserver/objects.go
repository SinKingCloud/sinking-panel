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

// foreignKeyDefinition 生成原生外键约束，引用对象全部通过标识符引用。
func (m *Manager) foreignKeyDefinition(key ForeignKeyDefinition) (string, error) {
	if len(key.Columns) == 0 || len(key.Columns) > 32 || len(key.Columns) != len(key.ReferenceColumns) {
		return "", errors.New("外键字段与引用字段必须一一对应，最多 32 个字段")
	}
	_, reference, err := m.target(key.ReferenceSchema, key.ReferenceTable)
	if err != nil {
		return "", err
	}
	columns, references := make([]string, 0, len(key.Columns)), make([]string, 0, len(key.ReferenceColumns))
	seen := make(map[string]bool, len(key.Columns))
	for index, column := range key.Columns {
		name, err := m.quote(column)
		if err != nil {
			return "", err
		}
		if seen[strings.ToLower(column)] {
			return "", errors.New("外键字段不能重复")
		}
		seen[strings.ToLower(column)] = true
		ref, err := m.quote(key.ReferenceColumns[index])
		if err != nil {
			return "", err
		}
		columns, references = append(columns, name), append(references, ref)
	}
	definition := ""
	if key.Name != "" {
		name, err := m.quote(key.Name)
		if err != nil || utf8.RuneCountInString(key.Name) > 128 {
			return "", errors.New("外键名称不合法，最多允许 128 个字符")
		}
		definition = "CONSTRAINT " + name + " "
	}
	definition += "FOREIGN KEY (" + strings.Join(columns, ", ") + ") REFERENCES " + reference + " (" + strings.Join(references, ", ") + ")"
	for _, action := range []struct {
		name  string
		value string
	}{{"DELETE", key.OnDelete}, {"UPDATE", key.OnUpdate}} {
		value := strings.ToUpper(strings.TrimSpace(action.value))
		if value == "" || value == "RESTRICT" {
			value = "NO ACTION"
		}
		switch value {
		case "NO ACTION", "CASCADE", "SET NULL", "SET DEFAULT":
		default:
			return "", errors.New("外键更新或删除动作不合法")
		}
		definition += " ON " + action.name + " " + value
	}
	return definition, nil
}

// CreateForeignKey 为用户表添加真实外键并校验已有数据。
func (m *Manager) CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	definition, err := m.foreignKeyDefinition(key)
	if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" WITH CHECK ADD "+definition); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// DropForeignKey 只删除属于当前用户表的外键约束。
func (m *Manager) DropForeignKey(ctx context.Context, schemaName, tableName, name string) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	constraint, err := m.quote(name)
	if err != nil {
		return nil, err
	}
	var found int
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys.foreign_keys WHERE parent_object_id = OBJECT_ID(@p1) AND name = @p2", table, name).Scan(&found); err != nil {
		return nil, err
	}
	if found != 1 {
		return nil, errors.New("当前表中不存在该外键")
	}
	if _, err = m.db.ExecContext(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+constraint); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// Objects 返回视图、存储过程/函数、触发器或外键的真实目录信息。
func (m *Manager) Objects(ctx context.Context, schemaName, kind string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	filter := ""
	switch kind {
	case "view":
		filter = "o.type = 'V'"
	case "routine":
		filter = "o.type IN ('P', 'PC', 'FN', 'IF', 'TF', 'FS', 'FT')"
	case "trigger":
		filter = "o.type IN ('TR', 'TA')"
	case "foreign_key":
	default:
		return nil, errors.New("对象类型不合法")
	}
	if kind != "foreign_key" {
		// 局部临时存储过程也受会话限制，不作为可跨请求管理的对象返回。
		query := "SELECT o.name, s.name AS [schema], o.type_desc AS type, p.name AS [table] FROM sys.objects o JOIN sys.schemas s ON s.schema_id = o.schema_id LEFT JOIN sys.objects p ON p.object_id = o.parent_object_id WHERE o.is_ms_shipped = 0 AND (o.name NOT LIKE '#%' OR o.name LIKE '##%') AND " + filter
		var args []interface{}
		if schemaName != "" {
			query += " AND s.name = @p1"
			args = append(args, schemaName)
		}
		return m.records(ctx, query+" ORDER BY s.name, o.name", args...)
	}
	query := "SELECT f.name, s.name AS [schema], t.name AS [table], c.name AS column_name, rs.name AS reference_schema, rt.name AS reference_table, rc.name AS reference_column, f.delete_referential_action_desc AS on_delete, f.update_referential_action_desc AS on_update, f.is_disabled, f.is_not_trusted FROM sys.foreign_keys f JOIN sys.tables t ON t.object_id = f.parent_object_id JOIN sys.schemas s ON s.schema_id = t.schema_id JOIN sys.foreign_key_columns fc ON fc.constraint_object_id = f.object_id JOIN sys.columns c ON c.object_id = fc.parent_object_id AND c.column_id = fc.parent_column_id JOIN sys.tables rt ON rt.object_id = f.referenced_object_id JOIN sys.schemas rs ON rs.schema_id = rt.schema_id JOIN sys.columns rc ON rc.object_id = fc.referenced_object_id AND rc.column_id = fc.referenced_column_id WHERE f.is_ms_shipped = 0"
	var args []interface{}
	if schemaName != "" {
		query += " AND s.name = @p1"
		args = append(args, schemaName)
	}
	records, err := m.records(ctx, query+" ORDER BY s.name, f.name, fc.constraint_column_id", args...)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0)
	positions := make(map[string]int)
	for _, record := range records {
		id := fmt.Sprint(record["schema"]) + "\x00" + fmt.Sprint(record["name"])
		position, found := positions[id]
		if !found {
			position = len(result)
			positions[id] = position
			record["type"] = "foreign_key"
			record["columns"], record["reference_columns"] = []string{}, []string{}
			record["on_delete"] = strings.ReplaceAll(fmt.Sprint(record["on_delete"]), "_", " ")
			record["on_update"] = strings.ReplaceAll(fmt.Sprint(record["on_update"]), "_", " ")
			result = append(result, record)
		}
		result[position]["columns"] = append(result[position]["columns"].([]string), fmt.Sprint(record["column_name"]))
		result[position]["reference_columns"] = append(result[position]["reference_columns"].([]string), fmt.Sprint(record["reference_column"]))
	}
	for _, record := range result {
		delete(record, "column_name")
		delete(record, "reference_column")
	}
	return result, nil
}

// ObjectDefinition 返回原始模块定义，外键生成对应的 ALTER TABLE 语句。
func (m *Manager) ObjectDefinition(ctx context.Context, schemaName, kind, name string) (_ string, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if kind != "foreign_key" {
		schema, _, err := m.target(schemaName, name)
		if err != nil {
			return "", err
		}
		filter := ""
		switch kind {
		case "view":
			filter = "o.type = 'V'"
		case "routine":
			filter = "o.type IN ('P', 'PC', 'FN', 'IF', 'TF', 'FS', 'FT')"
		case "trigger":
			filter = "o.type IN ('TR', 'TA')"
		default:
			return "", errors.New("对象类型不合法")
		}
		var definition sql.NullString
		query := "SELECT OBJECT_DEFINITION(o.object_id) FROM sys.objects o JOIN sys.schemas s ON s.schema_id = o.schema_id WHERE s.name = @p1 AND o.name = @p2 AND o.is_ms_shipped = 0 AND " + filter
		if err = m.db.QueryRowContext(ctx, query, schema, name).Scan(&definition); errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("对象不存在或当前账户无权查看")
		} else if err != nil {
			return "", err
		}
		if !definition.Valid {
			return "", errors.New("对象定义已加密或当前账户无权查看")
		}
		if len(definition.String) > 1<<20 {
			return "", errors.New("结构定义超过 1 MB")
		}
		return definition.String, nil
	}
	data, err := m.Objects(ctx, schemaName, kind)
	if err != nil {
		return "", err
	}
	for _, record := range data.([]map[string]interface{}) {
		if !strings.EqualFold(fmt.Sprint(record["name"]), name) {
			continue
		}
		if kind == "foreign_key" {
			key := ForeignKeyDefinition{
				Name:             name,
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
			_, table, err := m.target(fmt.Sprint(record["schema"]), fmt.Sprint(record["table"]))
			if err != nil {
				return "", err
			}
			validation := "CHECK"
			if m.truth(record["is_not_trusted"]) || m.truth(record["is_disabled"]) {
				validation = "NOCHECK"
			}
			statement := "ALTER TABLE " + table + " WITH " + validation + " ADD " + definition + ";"
			if m.truth(record["is_disabled"]) {
				constraint, _ := m.quote(name)
				statement += "\nALTER TABLE " + table + " NOCHECK CONSTRAINT " + constraint + ";"
			}
			return statement, nil
		}
	}
	return "", errors.New("对象不存在或当前账户无权查看")
}

// SaveObject 使用 CREATE/ALTER 保存模块，保持原对象权限，不用删除重建。
func (m *Manager) SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	schema, target, err := m.target(schemaName, name)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(name) > 128 || strings.EqualFold(schema, "sys") || strings.EqualFold(schema, "INFORMATION_SCHEMA") {
		return nil, errors.New("不能使用系统模式或超过 128 字符的对象名称")
	}
	tokens, err := m.schemaTokens(definition)
	if err != nil {
		return nil, err
	}
	position := 1
	if len(tokens) < 4 || (tokens[0] != "CREATE" && tokens[0] != "ALTER") {
		return nil, errors.New("请填写对应对象的完整 CREATE 或 ALTER 定义")
	}
	if len(tokens) > 5 && tokens[0] == "CREATE" && tokens[1] == "OR" && tokens[2] == "ALTER" {
		position = 3
	}
	objectType := tokens[position]
	if (kind == "view" && objectType != "VIEW") || (kind == "trigger" && objectType != "TRIGGER") || (kind == "routine" && objectType != "PROCEDURE" && objectType != "PROC" && objectType != "FUNCTION") || (kind != "view" && kind != "trigger" && kind != "routine") {
		return nil, errors.New("定义的对象类型与当前选择不一致")
	}
	position++
	objectName := strings.TrimPrefix(tokens[position], "\x01")
	position++
	if position+1 < len(tokens) && tokens[position] == "." {
		if !strings.EqualFold(objectName, schema) {
			return nil, errors.New("定义中的模式与当前选择不一致")
		}
		objectName = strings.TrimPrefix(tokens[position+1], "\x01")
		position += 2
	}
	if !strings.EqualFold(objectName, name) || (position < len(tokens) && tokens[position] == ".") {
		return nil, errors.New("定义中的对象名称与当前对象不一致")
	}
	if kind == "view" {
		for index, token := range tokens[position:] {
			if token == ";" && position+index != len(tokens)-1 {
				return nil, errors.New("视图定义不能包含附加语句")
			}
			switch token {
			case "INSERT", "UPDATE", "DELETE", "MERGE", "CREATE", "ALTER", "DROP", "EXEC", "EXECUTE", "USE", "SET", "GRANT", "REVOKE", "DENY", "TRUNCATE", "GO":
				return nil, errors.New("视图定义不能包含附加语句")
			}
		}
	} else {
		depth, body, ended := 0, false, false
		for index, token := range tokens[position:] {
			if token == "GO" {
				return nil, errors.New("请只提交一个对象的定义，不要包含 GO 批次分隔")
			}
			if ended && token != ";" && token != "CATCH" {
				return nil, errors.New("对象定义不能包含附加语句")
			}
			if token == "BEGIN" {
				// BEGIN TRAN/TRANSACTION 属于过程内容，不是 BEGIN ... END 块。
				following := ""
				if position+index+1 < len(tokens) {
					following = tokens[position+index+1]
				}
				if following != "TRAN" && following != "TRANSACTION" && following != "DISTRIBUTED" {
					body = true
					depth++
				}
			} else if body && token == "CASE" {
				depth++
			} else if body && token == "END" {
				depth--
				ended = depth == 0
				if position+index+1 < len(tokens) && tokens[position+index+1] == "TRY" {
					ended = false
				}
			}
		}
		if body && !ended {
			return nil, errors.New("对象定义的 BEGIN ... END 不完整")
		}
	}
	var existing sql.NullString
	var system sql.NullBool
	err = m.db.QueryRowContext(ctx, "SELECT type, is_ms_shipped FROM sys.objects WHERE object_id = OBJECT_ID(@p1)", target).Scan(&existing, &system)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if system.Valid && system.Bool {
		return nil, errors.New("不能修改 SqlServer 系统对象")
	}
	if existing.Valid {
		typeName := strings.TrimSpace(existing.String)
		if (kind == "view" && typeName != "V") || (kind == "trigger" && typeName != "TR") || (kind == "routine" && typeName != "P" && typeName != "FN" && typeName != "IF" && typeName != "TF") {
			return nil, errors.New("已有对象类型与当前选择不一致")
		}
		if !replace {
			return nil, errors.New("对象已存在，请通过编辑保存修改")
		}
	}
	// 旧版本也支持 CREATE/ALTER；避免依赖较新的 CREATE OR ALTER 语法。
	verb := "CREATE"
	if existing.Valid {
		verb = "ALTER"
	}
	identifier := `(?:\[(?:[^\]]|\]\])+\]|"(?:""|[^"])+"|[\p{L}\p{N}_@#$]+)`
	header := regexp.MustCompile(`(?is)^(\s*(?:(?:--[^\n]*(?:\n|$)|/\*.*?\*/)\s*)*)(?:CREATE\s+OR\s+ALTER|CREATE|ALTER)\s+(?:VIEW|FUNCTION|PROCEDURE|PROC|TRIGGER)\s+` + identifier + `(?:\s*\.\s*` + identifier + `)?`)
	if !header.MatchString(definition) {
		return nil, errors.New("对象定义请直接以 CREATE 或 ALTER 开头")
	}
	definition = header.ReplaceAllStringFunc(definition, func(match string) string {
		return header.FindStringSubmatch(match)[1] + verb + " " + objectType + " " + target
	})
	if _, err = m.db.ExecContext(ctx, definition); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// DropObject 依对象的真实类型删除模块，拒绝系统对象与类型不匹配。
func (m *Manager) DropObject(ctx context.Context, schemaName, kind, name string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	schema, target, err := m.target(schemaName, name)
	if err != nil {
		return nil, err
	}
	var objectType string
	var system bool
	if err = m.db.QueryRowContext(ctx, "SELECT type, is_ms_shipped FROM sys.objects WHERE object_id = OBJECT_ID(@p1)", target).Scan(&objectType, &system); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("对象不存在或当前账户无权访问")
	} else if err != nil {
		return nil, err
	}
	if system || strings.EqualFold(schema, "sys") || strings.EqualFold(schema, "INFORMATION_SCHEMA") {
		return nil, errors.New("不能删除 SqlServer 系统对象")
	}
	command := ""
	switch strings.TrimSpace(objectType) {
	case "V":
		if kind == "view" {
			command = "VIEW"
		}
	case "P", "PC":
		if kind == "routine" {
			command = "PROCEDURE"
		}
	case "FN", "IF", "TF", "FS", "FT":
		if kind == "routine" {
			command = "FUNCTION"
		}
	case "TR", "TA":
		if kind == "trigger" {
			command = "TRIGGER"
		}
	}
	if command == "" {
		return nil, errors.New("对象类型与当前选择不一致")
	}
	if _, err = m.db.ExecContext(ctx, "DROP "+command+" "+target); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// SelectObjects 按名称搜索并分页查询对象摘要；定义在详情接口按需读取。
func (m *Manager) SelectObjects(ctx context.Context, schema, kind, keyword string, pageNumber, pageSize int) (interface{}, error) {
	if kind == "foreign_key" {
		return m.Objects(ctx, schema, kind)
	}
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageNumber > 1000000 || pageSize > 1000 {
		return nil, errors.New("分页参数不合法")
	}

	if schema == "" {
		schema = "dbo"
	}
	filter := ""
	switch kind {
	case "view":
		filter = "o.type='V'"
	case "routine":
		filter = "o.type IN ('P','PC','FN','IF','TF','FS','FT')"
	case "trigger":
		filter = "o.type IN ('TR','TA')"
	default:
		return nil, errors.New("对象类型不合法")
	}
	from := " FROM sys.objects o JOIN sys.schemas s ON s.schema_id=o.schema_id LEFT JOIN sys.objects p ON p.object_id=o.parent_object_id WHERE o.is_ms_shipped=0 AND (o.name NOT LIKE '#%' OR o.name LIKE '##%') AND " + filter + " AND s.name=@p1 AND o.name LIKE @p2"
	args := []interface{}{schema, "%" + keyword + "%"}
	var total int64
	if err := m.db.QueryRowContext(ctx, "SELECT COUNT(*)"+from, args...).Scan(&total); err != nil {
		return nil, m.safeError(err)
	}
	if total > 0 && int64((pageNumber-1)*pageSize) >= total {
		pageNumber = int((total-1)/int64(pageSize)) + 1
	}
	args = append(args, (pageNumber-1)*pageSize, pageNumber*pageSize)
	// ROW_NUMBER 兼容 SQL Server 2005 及以上，无需 OFFSET/FETCH。
	query := "SELECT name,[schema],type,[table] FROM (SELECT o.name,s.name AS [schema],o.type_desc AS type,p.name AS [table],ROW_NUMBER() OVER(ORDER BY s.name,o.name,o.object_id) AS row_number" + from + ") AS objects WHERE row_number>@p3 AND row_number<=@p4 ORDER BY row_number"
	list, err := m.records(ctx, query, args...)
	if err != nil {
		return nil, m.safeError(err)
	}
	return map[string]interface{}{
		"list":      list,
		"total":     total,
		"page":      pageNumber,
		"page_size": pageSize,
	}, nil
}
