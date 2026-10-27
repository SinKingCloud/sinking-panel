package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// foreignKeyDefinition 生成建表使用的外键约束，不跨数据库引用。
func (m *Manager) foreignKeyDefinition(key ForeignKeyDefinition) (string, error) {
	if len(key.Columns) == 0 || len(key.Columns) > 2000 || len(key.Columns) != len(key.ReferenceColumns) {
		return "", errors.New("外键字段与引用字段必须一一对应")
	}
	if key.ReferenceSchema != "" && key.ReferenceSchema != "main" {
		return "", errors.New("Sqlite 外键不能引用其他数据库")
	}
	reference, err := m.quote(key.ReferenceTable)
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
		if err != nil {
			return "", err
		}
		definition = "CONSTRAINT " + name + " "
	}
	definition += "FOREIGN KEY (" + strings.Join(columns, ", ") + ") REFERENCES " + reference + " (" + strings.Join(references, ", ") + ")"
	for _, action := range []struct {
		name  string
		value string
	}{{"DELETE", key.OnDelete}, {"UPDATE", key.OnUpdate}} {
		value := strings.ToUpper(strings.TrimSpace(action.value))
		if value == "" {
			value = "NO ACTION"
		}
		switch value {
		case "NO ACTION", "RESTRICT", "CASCADE", "SET NULL", "SET DEFAULT":
		default:
			return "", errors.New("外键更新或删除动作不合法")
		}
		definition += " ON " + action.name + " " + value
	}
	return definition, nil
}

// CreateForeignKey 通过事务迁移添加外键，完整保留原数据与其他约束。
func (m *Manager) CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (*Result, error) {
	return m.rebuildTable(ctx, schemaName, tableName, nil, "", &key, false)
}

// DropForeignKey 删除指定外键约束，其他依赖与数据在同一事务保留。
func (m *Manager) DropForeignKey(ctx context.Context, schemaName, tableName, name string) (*Result, error) {
	return m.rebuildTable(ctx, schemaName, tableName, nil, name, nil, true)
}

// Objects 返回视图、触发器或外键；Sqlite 不支持存储过程。
func (m *Manager) Objects(ctx context.Context, schemaName, kind string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if schemaName != "" && schemaName != "main" {
		return nil, errors.New("Sqlite 仅支持 main 数据库")
	}
	if kind == "routine" {
		return nil, errors.New("Sqlite 不支持存储过程或持久化函数")
	}
	if kind == "view" || kind == "trigger" {
		return m.records(ctx, "SELECT name, 'main' AS schema, type, tbl_name AS 'table' FROM main.sqlite_master WHERE type = ? AND name NOT GLOB 'sqlite_*' ORDER BY name", kind)
	}
	if kind != "foreign_key" {
		return nil, errors.New("对象类型不合法")
	}
	tables, err := m.records(ctx, "SELECT name FROM main.sqlite_master WHERE type = 'table' AND name NOT GLOB 'sqlite_*' ORDER BY name")
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0)
	for _, table := range tables {
		tableName := fmt.Sprint(table["name"])
		quoted, _ := m.quote(tableName)
		keys, err := m.records(ctx, "PRAGMA main.foreign_key_list("+quoted+")")
		if err != nil {
			return nil, err
		}
		names := make([]string, 0)
		if len(keys) > 0 {
			var definition string
			if err = m.db.QueryRowContext(ctx, "SELECT sql FROM main.sqlite_master WHERE type = 'table' AND name = ? COLLATE NOCASE", tableName).Scan(&definition); err != nil {
				return nil, err
			}
			tokens, err := m.schemaTokens(definition)
			if err != nil {
				return nil, err
			}
			depth, clauseStart := 0, -1
			for index, token := range tokens {
				if token == "(" {
					depth++
					if depth == 1 {
						clauseStart = index + 1
					}
				} else if token == ")" {
					depth--
				} else if token == "," && depth == 1 {
					clauseStart = index + 1
				} else if token == "REFERENCES" && depth == 1 {
					name := ""
					if clauseStart >= 0 && clauseStart+2 < len(tokens) && tokens[clauseStart] == "CONSTRAINT" && tokens[clauseStart+2] == "FOREIGN" {
						name = strings.TrimPrefix(tokens[clauseStart+1], "\x01")
					} else if index >= 2 && tokens[index-2] == "CONSTRAINT" {
						name = strings.TrimPrefix(tokens[index-1], "\x01")
					}
					names = append(names, name)
				}
			}
		}
		positions := make(map[string]int)
		for _, key := range keys {
			id := fmt.Sprint(key["id"])
			position, found := positions[id]
			if !found {
				if len(result) >= 10000 {
					return nil, errors.New("外键数量超过单次读取上限")
				}
				position = len(result)
				positions[id] = position
				identity := tableName + "." + id
				name := identity
				if number, parseErr := strconv.Atoi(id); parseErr == nil && number >= 0 && number < len(names) && names[len(names)-number-1] != "" {
					name = names[len(names)-number-1]
				}
				result = append(result, map[string]interface{}{
					"name":              name,
					"identity":          identity,
					"schema":            "main",
					"type":              "foreign_key",
					"table":             tableName,
					"columns":           []string{},
					"reference_schema":  "main",
					"reference_table":   key["table"],
					"reference_columns": []string{},
					"on_delete":         key["on_delete"],
					"on_update":         key["on_update"],
				})
			}
			result[position]["columns"] = append(result[position]["columns"].([]string), fmt.Sprint(key["from"]))
			reference := ""
			if key["to"] != nil {
				reference = fmt.Sprint(key["to"])
			}
			result[position]["reference_columns"] = append(result[position]["reference_columns"].([]string), reference)
		}
	}
	return result, nil
}

// ObjectDefinition 返回对象的原始定义；外键返回所在表的完整建表定义。
func (m *Manager) ObjectDefinition(ctx context.Context, schemaName, kind, name string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if kind == "foreign_key" {
		data, err := m.Objects(ctx, schemaName, kind)
		if err != nil {
			return "", err
		}
		for _, record := range data.([]map[string]interface{}) {
			if strings.EqualFold(fmt.Sprint(record["name"]), name) || fmt.Sprint(record["identity"]) == name {
				return m.DDL(ctx, schemaName, fmt.Sprint(record["table"]))
			}
		}
		return "", errors.New("外键不存在")
	}
	if kind != "view" && kind != "trigger" {
		return "", errors.New("Sqlite 只支持视图和触发器定义")
	}
	if _, err := m.target(schemaName, name); err != nil {
		return "", err
	}
	var definition string
	err := m.db.QueryRowContext(ctx, "SELECT sql FROM main.sqlite_master WHERE type = ? AND name = ? COLLATE NOCASE", kind, name).Scan(&definition)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("对象不存在")
	}
	if len(definition) > 1<<20 {
		return "", errors.New("结构定义超过 1 MB")
	}
	return definition, err
}

// SaveObject 创建或原子替换视图/触发器，校验定义中的对象名称与请求相同。
func (m *Manager) SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	if kind != "view" && kind != "trigger" {
		return nil, errors.New("Sqlite 仅支持保存视图和触发器")
	}
	target, err := m.target(schemaName, name)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.ToLower(name), "sqlite_") {
		return nil, errors.New("不能修改 Sqlite 系统对象")
	}
	tokens, err := m.schemaTokens(definition)
	if err != nil {
		return nil, err
	}
	if len(tokens) < 4 || tokens[0] != "CREATE" || tokens[1] != strings.ToUpper(kind) {
		return nil, errors.New("请填写对应对象的完整 CREATE 定义")
	}
	position := 2
	if len(tokens) > 5 && tokens[2] == "IF" && tokens[3] == "NOT" && tokens[4] == "EXISTS" {
		position = 5
	}
	objectName := strings.TrimPrefix(tokens[position], "\x01")
	position++
	if position+1 < len(tokens) && tokens[position] == "." {
		if !strings.EqualFold(objectName, "main") {
			return nil, errors.New("对象定义只能使用 main 数据库")
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
		}
	} else {
		depth, body, ended := 0, false, false
		for _, token := range tokens[position:] {
			if ended && token != ";" {
				return nil, errors.New("触发器定义不能包含附加语句")
			}
			if token == "BEGIN" {
				body = true
				depth++
			} else if body && token == "CASE" {
				depth++
			} else if body && token == "END" {
				depth--
				ended = depth == 0
			}
		}
		if !body || !ended {
			return nil, errors.New("触发器必须包含完整的 BEGIN ... END 定义")
		}
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if replace {
		if _, err = tx.ExecContext(ctx, "DROP "+strings.ToUpper(kind)+" IF EXISTS "+target); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, definition); err != nil {
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

// DropObject 删除用户视图或触发器，删除前确认对象类型。
func (m *Manager) DropObject(ctx context.Context, schemaName, kind, name string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	if kind != "view" && kind != "trigger" {
		return nil, errors.New("Sqlite 仅支持删除视图和触发器")
	}
	target, err := m.target(schemaName, name)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.ToLower(name), "sqlite_") {
		return nil, errors.New("不能删除 Sqlite 系统对象")
	}
	var existing string
	err = m.db.QueryRowContext(ctx, "SELECT type FROM main.sqlite_master WHERE name = ? COLLATE NOCASE", name).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("对象不存在")
	} else if err != nil {
		return nil, err
	}
	if existing != kind {
		return nil, errors.New("对象类型与当前选择不一致")
	}
	if _, err = m.db.ExecContext(ctx, "DROP "+strings.ToUpper(kind)+" "+target); err != nil {
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

	if schema != "" && schema != "main" {
		return nil, errors.New("仅支持 main 数据库")
	}
	if kind != "view" && kind != "trigger" {
		return nil, errors.New("当前数据库不支持该对象类型")
	}
	from := " FROM main.sqlite_master WHERE type=? AND name NOT GLOB 'sqlite_*' AND name LIKE ?"
	args := []interface{}{kind, "%" + keyword + "%"}
	var total int64
	if err := m.db.QueryRowContext(ctx, "SELECT COUNT(*)"+from, args...).Scan(&total); err != nil {
		return nil, err
	}
	if total > 0 && int64((pageNumber-1)*pageSize) >= total {
		pageNumber = int((total-1)/int64(pageSize)) + 1
	}
	args = append(args, pageSize, (pageNumber-1)*pageSize)
	list, err := m.records(ctx, "SELECT name,'main' AS schema,type,tbl_name AS 'table'"+from+" ORDER BY name LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":      list,
		"total":     total,
		"page":      pageNumber,
		"page_size": pageSize,
	}, nil
}
