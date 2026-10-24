package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Objects 返回 Postgres 业务模式中的视图、程序、触发器和外键。
func (m *Manager) Objects(ctx context.Context, schemaName, kind string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema := schemaName
	if schema == "" {
		schema = "public"
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") {
		return nil, errors.New("请选择业务模式")
	}
	switch kind {
	case "view":
		return m.records(ctx, "SELECT c.relname AS name, n.nspname AS schema, CASE WHEN c.relkind = 'm' THEN 'MATERIALIZED VIEW' ELSE 'VIEW' END AS type, pg_get_viewdef(c.oid, true) AS definition FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relkind IN ('v','m') ORDER BY c.relname", schema)
	case "routine":
		routineType, routineWhere, catalogErr := m.routineCatalog(ctx)
		if catalogErr != nil {
			return nil, catalogErr
		}
		return m.records(ctx, "SELECT p.proname AS name, n.nspname AS schema, "+routineType+" AS type, p.proname||'('||pg_get_function_identity_arguments(p.oid)||')' AS identity, pg_get_functiondef(p.oid) AS definition FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = $1 AND "+routineWhere+" ORDER BY p.proname,p.oid", schema)
	case "trigger":
		version, err := m.serverVersion(ctx)
		if err != nil {
			return nil, err
		}
		definition, visible := "pg_get_triggerdef(t.oid, true)", "NOT t.tgisinternal"
		if version < 90000 {
			definition, visible = "pg_get_triggerdef(t.oid)", "t.tgconstraint = 0"
		}
		return m.records(ctx, "SELECT t.tgname AS name, n.nspname AS schema, 'TRIGGER' AS type, c.relname AS table, c.relname||'.'||t.tgname AS identity, "+definition+" AS definition, t.tgenabled AS enabled, obj_description(t.oid, 'pg_trigger') AS comment FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND "+visible+" ORDER BY c.relname,t.tgname", schema)
	case "foreign_key":
		// 旧版不支持 FROM 函数引用同级列；按服务器索引容量展开，不依赖 LATERAL。
		records, queryErr := m.records(ctx, "SELECT p.conname AS name, n.nspname AS schema, c.relname AS table, c.relname||'.'||p.conname AS identity, rn.nspname AS reference_schema, r.relname AS reference_table, a.attname AS column_name, ra.attname AS reference_column, p.confdeltype AS on_delete, p.confupdtype AS on_update, pg_get_constraintdef(p.oid, true) AS definition FROM pg_constraint p JOIN pg_class c ON c.oid = p.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace JOIN pg_class r ON r.oid = p.confrelid JOIN pg_namespace rn ON rn.oid = r.relnamespace CROSS JOIN generate_series(1, current_setting('max_index_keys')::int) AS k(position) JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = p.conkey[k.position] JOIN pg_attribute ra ON ra.attrelid = r.oid AND ra.attnum = p.confkey[k.position] WHERE n.nspname = $1 AND p.contype = 'f' AND k.position <= array_upper(p.conkey, 1) ORDER BY c.relname,p.conname,k.position", schema)
		if queryErr != nil {
			return nil, queryErr
		}
		result := make([]map[string]interface{}, 0)
		positions := make(map[string]int)
		rules := map[string]string{"a": "NO ACTION", "r": "RESTRICT", "c": "CASCADE", "n": "SET NULL", "d": "SET DEFAULT"}
		for _, record := range records {
			identity := fmt.Sprint(record["identity"])
			position, ok := positions[identity]
			if !ok {
				position = len(result)
				positions[identity] = position
				result = append(result, map[string]interface{}{"name": record["name"], "schema": record["schema"], "type": "FOREIGN KEY", "identity": identity, "table": record["table"], "definition": record["definition"], "columns": []string{}, "reference_schema": record["reference_schema"], "reference_table": record["reference_table"], "reference_columns": []string{}, "on_delete": rules[fmt.Sprint(record["on_delete"])], "on_update": rules[fmt.Sprint(record["on_update"])]})
			}
			result[position]["columns"] = append(result[position]["columns"].([]string), fmt.Sprint(record["column_name"]))
			result[position]["reference_columns"] = append(result[position]["reference_columns"].([]string), fmt.Sprint(record["reference_column"]))
		}
		return result, nil
	default:
		return nil, errors.New("对象类型仅支持 view、routine、trigger 或 foreign_key")
	}
}

// ObjectDefinition 返回指定对象的真实 DDL；重载程序和同名触发器通过 identity 定位。
func (m *Manager) ObjectDefinition(ctx context.Context, schemaName, kind, name string) (out string, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema := schemaName
	if schema == "" {
		schema = "public"
	}
	switch kind {
	case "view":
		return m.DDL(ctx, schema, name)
	case "routine":
		_, routineWhere, catalogErr := m.routineCatalog(ctx)
		if catalogErr != nil {
			return "", catalogErr
		}
		records, queryErr := m.records(ctx, "SELECT pg_get_functiondef(p.oid) AS definition FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = $1 AND (p.proname = $2 OR p.proname||'('||pg_get_function_identity_arguments(p.oid)||')' = $2) AND "+routineWhere+"", schema, name)
		if queryErr != nil {
			return "", queryErr
		}
		if len(records) == 0 {
			return "", errors.New("存储程序不存在或无权查看")
		}
		if len(records) > 1 {
			return "", errors.New("存在同名重载程序，请选择包含参数签名的对象")
		}
		return fmt.Sprint(records[0]["definition"]), nil
	case "trigger", "foreign_key":
		objects, queryErr := m.Objects(ctx, schema, kind)
		if queryErr != nil {
			return "", queryErr
		}
		matches := make([]map[string]interface{}, 0)
		for _, record := range objects.([]map[string]interface{}) {
			if record["name"] == name || record["identity"] == name {
				matches = append(matches, record)
			}
		}
		if len(matches) == 0 {
			return "", errors.New("对象不存在或无权查看")
		}
		if len(matches) > 1 {
			return "", errors.New("当前模式存在同名对象，请选择包含所属表的对象")
		}
		record := matches[0]
		definition := fmt.Sprint(record["definition"])
		if kind == "foreign_key" {
			_, table, _ := m.target(schema, fmt.Sprint(record["table"]))
			definition = "ALTER TABLE " + table + " ADD CONSTRAINT "
			constraint, _ := m.quote(fmt.Sprint(record["name"]))
			definition += constraint + " " + fmt.Sprint(record["definition"])
		}
		return strings.TrimSuffix(strings.TrimSpace(definition), ";") + ";", nil
	default:
		return "", errors.New("不支持的对象类型")
	}
}

// objectSQL 只允许修改选择对象的一条 DDL，函数 dollar quote 正文作为整体处理。
func (m *Manager) objectSQL(schemaName, kind, name, definition string) (string, string, error) {
	definition = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(definition), ";"))
	if definition == "" || len(definition) > 1<<20 || strings.ContainsRune(definition, 0) {
		return "", "", errors.New("对象定义不能为空且不能超过 1 MB")
	}
	identifier := `(?:"(?:[^"]|"")+"|[a-zA-Z_][a-zA-Z0-9_$]*)`
	match := regexp.MustCompile(`(?is)^(CREATE(?:\s+OR\s+REPLACE)?|ALTER)\s+(MATERIALIZED\s+VIEW|VIEW|(?:CONSTRAINT\s+)?TRIGGER|FUNCTION|PROCEDURE)\s+(` + identifier + `)(?:\s*\.\s*(` + identifier + `))?(?:\s|\(|$)`).FindStringSubmatch(definition)
	if match == nil {
		return "", "", errors.New("请输入完整的 CREATE 或 ALTER 对象定义")
	}
	objectType := strings.TrimPrefix(strings.ToUpper(strings.Join(strings.Fields(match[2]), " ")), "CONSTRAINT ")
	valid := (kind == "view" && (objectType == "VIEW" || objectType == "MATERIALIZED VIEW")) || (kind == "trigger" && objectType == "TRIGGER") || (kind == "routine" && (objectType == "FUNCTION" || objectType == "PROCEDURE"))
	if !valid {
		return "", "", errors.New("对象定义与选择的对象类型不一致")
	}
	decode := func(value string) string {
		if strings.HasPrefix(value, `"`) {
			return strings.ReplaceAll(value[1:len(value)-1], `""`, `"`)
		}
		return strings.ToLower(value)
	}
	targetName, targetSchema := decode(match[3]), schemaName
	if match[4] != "" {
		targetSchema, targetName = decode(match[3]), decode(match[4])
	}
	if targetSchema == "" {
		targetSchema = "public"
	}
	if kind == "routine" {
		if position := strings.Index(name, "("); position >= 0 {
			name = name[:position]
		}
	} else if kind == "trigger" {
		if position := strings.LastIndex(name, "."); position >= 0 {
			name = name[position+1:]
		}
	}
	if targetSchema != schemaName || targetName != name {
		return "", "", errors.New("对象定义必须使用当前选择的模式和对象名称")
	}
	quote, depth, blocks := byte(0), 0, 0
	dollar := ""
	word := strings.Builder{}
	flush := func() {
		token := strings.ToUpper(word.String())
		word.Reset()
		if token == "BEGIN" || token == "CASE" {
			blocks++
		} else if token == "END" && blocks > 0 {
			blocks--
		}
	}
	for index := 0; index < len(definition); index++ {
		ch := definition[index]
		if dollar != "" {
			if strings.HasPrefix(definition[index:], dollar) {
				index += len(dollar) - 1
				dollar = ""
			}
			continue
		}
		if quote != 0 {
			if ch == '\\' && quote == '\'' {
				index++
				continue
			}
			if ch == quote {
				if index+1 < len(definition) && definition[index+1] == quote {
					index++
				} else {
					quote = 0
				}
			}
			continue
		}
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
			word.WriteByte(ch)
			continue
		}
		flush()
		switch ch {
		case '\'', '"':
			quote = ch
		case '$':
			if match := regexp.MustCompile(`^\$(?:[a-zA-Z_][a-zA-Z0-9_]*)?\$`).FindString(definition[index:]); match != "" {
				dollar = match
				index += len(match) - 1
			}
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return "", "", errors.New("对象定义括号不匹配")
			}
		case ';':
			if blocks == 0 {
				return "", "", errors.New("只允许提交一条对象定义")
			}
		case '-', '/':
			if index+1 < len(definition) && ch == '-' && definition[index+1] == '-' {
				if end := strings.IndexByte(definition[index:], '\n'); end >= 0 {
					index += end
				} else {
					index = len(definition)
				}
			} else if index+1 < len(definition) && ch == '/' && definition[index+1] == '*' {
				nested := 1
				index += 2
				for index < len(definition) && nested > 0 {
					if strings.HasPrefix(definition[index:], "/*") {
						nested++
						index += 2
					} else if strings.HasPrefix(definition[index:], "*/") {
						nested--
						index += 2
					} else {
						index++
					}
				}
				if nested != 0 {
					return "", "", errors.New("对象定义注释未结束")
				}
				index--
			}
		}
	}
	flush()
	if quote != 0 || depth != 0 || dollar != "" || blocks != 0 {
		return "", "", errors.New("对象定义的引号、括号或 dollar quote 不匹配")
	}
	return definition, objectType, nil
}

// SaveObject 保存真实对象，函数、普通视图与新版本触发器使用原生 CREATE OR REPLACE。
func (m *Manager) SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema := schemaName
	if schema == "" {
		schema = "public"
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") || m.database == "template0" || m.database == "template1" {
		return nil, errors.New("不能修改系统模式中的对象")
	}
	statement, objectType, err := m.objectSQL(schema, kind, name, definition)
	if err != nil {
		return nil, err
	}
	var trigger map[string]interface{}
	if replace && kind == "trigger" && strings.HasPrefix(strings.ToUpper(statement), "CREATE") {
		objects, queryErr := m.Objects(ctx, schema, "trigger")
		if queryErr != nil {
			return nil, queryErr
		}
		for _, record := range objects.([]map[string]interface{}) {
			if record["name"] == name || record["identity"] == name {
				if trigger != nil {
					return nil, errors.New("存在同名触发器，请选择包含所属表的对象")
				}
				trigger = record
			}
		}
		if trigger == nil {
			return nil, errors.New("触发器不存在")
		}
		// 触发器名称不带 schema，所属表由 ON 子句限定；更新不允许移到另一张表。
		identifier := `(?:"(?:[^"]|"")+"|[a-zA-Z_][a-zA-Z0-9_$]*)`
		matches := regexp.MustCompile(`(?is)\bON\s+(` + identifier + `)(?:\s*\.\s*(` + identifier + `))?\s+`).FindStringSubmatch(statement)
		if matches == nil {
			return nil, errors.New("触发器定义缺少所属表")
		}
		decode := func(value string) string {
			if strings.HasPrefix(value, `"`) {
				return strings.ReplaceAll(value[1:len(value)-1], `""`, `"`)
			}
			return strings.ToLower(value)
		}
		targetSchema, targetTable := schema, decode(matches[1])
		if matches[2] != "" {
			targetSchema, targetTable = decode(matches[1]), decode(matches[2])
		}
		if targetSchema != schema || targetTable != fmt.Sprint(trigger["table"]) {
			return nil, errors.New("触发器定义必须保留当前所属表")
		}
		statement = regexp.MustCompile(`(?i)^CREATE\s+OR\s+REPLACE\s+`).ReplaceAllString(statement, "CREATE ")
	}
	if replace && objectType != "TRIGGER" && strings.HasPrefix(strings.ToUpper(statement), "CREATE") && !strings.HasPrefix(strings.ToUpper(statement), "CREATE OR REPLACE") {
		if objectType == "MATERIALIZED VIEW" {
			return nil, errors.New("Postgres 不支持原子替换物化视图定义，请使用新名称创建")
		}
		statement = "CREATE OR REPLACE" + statement[len("CREATE"):]
	}
	// 使用单独事务设置 search_path，非限定对象名不会误写其他模式，连接池设置也不会泄漏。
	quoted, _ := m.quote(schema)
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "SET LOCAL search_path = "+quoted+", pg_catalog"); err != nil {
		return nil, err
	}
	if trigger != nil {
		_, table, _ := m.target(schema, fmt.Sprint(trigger["table"]))
		triggerName, _ := m.quote(fmt.Sprint(trigger["name"]))
		if _, err = tx.ExecContext(ctx, "DROP TRIGGER "+triggerName+" ON "+table); err != nil {
			return nil, err
		}
	}
	started := time.Now()
	result, err := tx.ExecContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if trigger != nil {
		_, table, _ := m.target(schema, fmt.Sprint(trigger["table"]))
		triggerName, _ := m.quote(fmt.Sprint(trigger["name"]))
		if mode := map[string]string{"D": "DISABLE", "A": "ENABLE ALWAYS", "R": "ENABLE REPLICA"}[fmt.Sprint(trigger["enabled"])]; mode != "" {
			if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" "+mode+" TRIGGER "+triggerName); err != nil {
				return nil, err
			}
		}
		if trigger["comment"] != nil {
			comment := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(trigger["comment"]), "\\", "\\\\"), "'", "''")
			if _, err = tx.ExecContext(ctx, "COMMENT ON TRIGGER "+triggerName+" ON "+table+" IS E'"+comment+"'"); err != nil {
				return nil, err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, Affected: affected, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// DropObject 删除指定业务对象，重载程序及同名触发器通过目录信息精确定位。
func (m *Manager) DropObject(ctx context.Context, schemaName, kind, name string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema := schemaName
	if schema == "" {
		schema = "public"
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") || m.database == "template0" || m.database == "template1" {
		return nil, errors.New("不能删除系统模式中的对象")
	}
	switch kind {
	case "view":
		_, view, quoteErr := m.target(schema, name)
		if quoteErr != nil {
			return nil, quoteErr
		}
		var viewKind string
		err = m.db.QueryRowContext(ctx, "SELECT c.relkind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind IN ('v','m')", schema, name).Scan(&viewKind)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("视图不存在")
		}
		if err != nil {
			return nil, err
		}
		clause := "DROP VIEW "
		if viewKind == "m" {
			clause = "DROP MATERIALIZED VIEW "
		}
		return m.executeMaintenance(ctx, clause+view)
	case "routine":
		routineType, routineWhere, catalogErr := m.routineCatalog(ctx)
		if catalogErr != nil {
			return nil, catalogErr
		}
		records, queryErr := m.records(ctx, "SELECT quote_ident(n.nspname)||'.'||quote_ident(p.proname)||'('||pg_get_function_identity_arguments(p.oid)||')' AS target, "+routineType+" AS type FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = $1 AND (p.proname = $2 OR p.proname||'('||pg_get_function_identity_arguments(p.oid)||')' = $2) AND "+routineWhere+"", schema, name)
		if queryErr != nil {
			return nil, queryErr
		}
		if len(records) == 0 {
			return nil, errors.New("存储程序不存在")
		}
		if len(records) > 1 {
			return nil, errors.New("存在同名重载程序，请选择包含参数签名的对象")
		}
		return m.executeMaintenance(ctx, "DROP "+fmt.Sprint(records[0]["type"])+" "+fmt.Sprint(records[0]["target"]))
	case "trigger", "foreign_key":
		objects, queryErr := m.Objects(ctx, schema, kind)
		if queryErr != nil {
			return nil, queryErr
		}
		matches := make([]map[string]interface{}, 0)
		for _, record := range objects.([]map[string]interface{}) {
			if record["name"] == name || record["identity"] == name {
				matches = append(matches, record)
			}
		}
		if len(matches) == 0 {
			return nil, errors.New("对象不存在")
		}
		if len(matches) > 1 {
			return nil, errors.New("当前模式存在同名对象，请选择包含所属表的对象")
		}
		record := matches[0]
		if kind == "foreign_key" {
			return m.DropForeignKey(ctx, schema, fmt.Sprint(record["table"]), fmt.Sprint(record["name"]))
		}
		_, table, _ := m.target(schema, fmt.Sprint(record["table"]))
		trigger, _ := m.quote(fmt.Sprint(record["name"]))
		return m.executeMaintenance(ctx, "DROP TRIGGER "+trigger+" ON "+table)
	default:
		return nil, errors.New("不支持的对象类型")
	}
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
		schema = "public"
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") {
		return nil, errors.New("请选择业务模式")
	}
	var query string
	switch kind {
	case "view":
		query = "SELECT c.relname AS name,n.nspname AS schema,CASE WHEN c.relkind='m' THEN 'MATERIALIZED VIEW' ELSE 'VIEW' END AS type FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname=$1 AND c.relkind IN ('v','m')"
	case "routine":
		routineType, routineWhere, err := m.routineCatalog(ctx)
		if err != nil {
			return nil, err
		}
		query = "SELECT p.proname AS name,n.nspname AS schema," + routineType + " AS type,p.proname||'('||pg_get_function_identity_arguments(p.oid)||')' AS identity FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname=$1 AND " + routineWhere
	case "trigger":
		version, err := m.serverVersion(ctx)
		if err != nil {
			return nil, err
		}
		visible := "NOT t.tgisinternal"
		if version < 90000 {
			visible = "t.tgconstraint=0"
		}
		query = "SELECT t.tgname AS name,n.nspname AS schema,'TRIGGER' AS type,c.relname AS table,c.relname||'.'||t.tgname AS identity FROM pg_trigger t JOIN pg_class c ON c.oid=t.tgrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname=$1 AND " + visible
	default:
		return nil, errors.New("对象类型不合法")
	}
	query = "SELECT * FROM (" + query + ") AS objects WHERE name ILIKE $2"
	args := []interface{}{schema, "%" + keyword + "%"}
	var total int64
	if err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ("+query+") AS counted", args...).Scan(&total); err != nil {
		return nil, m.safeError(err)
	}
	if total > 0 && int64((pageNumber-1)*pageSize) >= total {
		pageNumber = int((total-1)/int64(pageSize)) + 1
	}
	args = append(args, pageSize, (pageNumber-1)*pageSize)
	order := "name"
	if kind == "routine" || kind == "trigger" {
		order += ",identity"
	}
	list, err := m.records(ctx, query+" ORDER BY "+order+" LIMIT $3 OFFSET $4", args...)
	if err != nil {
		return nil, m.safeError(err)
	}
	return map[string]interface{}{"list": list, "total": total, "page": pageNumber, "page_size": pageSize}, nil
}
