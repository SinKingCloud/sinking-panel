package mysql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"server/app/util/page"
	"strings"
	"time"
)

// Objects 分页列举当前库的对象；页码和容量均为 0 时供内部业务读取完整元数据，外键仍返回聚合数组。
func (m *Manager) Objects(ctx context.Context, schemaName, kind, keyword string, pageNumber, pageSize int) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if schemaName != "" && schemaName != m.database {
		return nil, errors.New("只能查看当前登记数据库的对象")
	}
	schemaName = m.database
	paginated := pageNumber != 0 || pageSize != 0
	if paginated && (pageNumber < 1 || pageNumber > 1000000 || pageSize < 1 || pageSize > 1000) {
		return nil, errors.New("页码应为 1 至 1000000，每页数量应为 1 至 1000")
	}
	var fields, source, nameColumn, order, definition string
	switch kind {
	case "view":
		fields = "TABLE_NAME AS name, TABLE_SCHEMA AS `schema`, 'VIEW' AS type"
		source = " FROM information_schema.VIEWS WHERE TABLE_SCHEMA = ?"
		nameColumn, order = "TABLE_NAME", "TABLE_NAME"
		definition = ", VIEW_DEFINITION AS definition"
	case "routine":
		fields = "ROUTINE_NAME AS name, ROUTINE_SCHEMA AS `schema`, ROUTINE_TYPE AS type"
		source = " FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = ?"
		nameColumn, order = "ROUTINE_NAME", "ROUTINE_NAME, ROUTINE_TYPE"
		definition = ", ROUTINE_DEFINITION AS definition"
	case "trigger":
		fields = "TRIGGER_NAME AS name, TRIGGER_SCHEMA AS `schema`, 'TRIGGER' AS type, EVENT_OBJECT_TABLE AS `table`"
		source = " FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA = ?"
		nameColumn, order = "TRIGGER_NAME", "TRIGGER_NAME"
		definition = ", ACTION_STATEMENT AS definition, ACTION_TIMING AS timing, EVENT_MANIPULATION AS event"
	case "foreign_key":
		records, queryErr := m.records(ctx, "SELECT k.CONSTRAINT_NAME AS name, k.CONSTRAINT_SCHEMA AS `schema`, k.TABLE_NAME AS `table`, k.COLUMN_NAME AS column_name, k.REFERENCED_TABLE_SCHEMA AS reference_schema, k.REFERENCED_TABLE_NAME AS reference_table, k.REFERENCED_COLUMN_NAME AS reference_column, r.DELETE_RULE AS on_delete, r.UPDATE_RULE AS on_update FROM information_schema.KEY_COLUMN_USAGE k JOIN information_schema.REFERENTIAL_CONSTRAINTS r ON r.CONSTRAINT_SCHEMA = k.CONSTRAINT_SCHEMA AND r.CONSTRAINT_NAME = k.CONSTRAINT_NAME AND r.TABLE_NAME = k.TABLE_NAME WHERE k.CONSTRAINT_SCHEMA = ? AND k.REFERENCED_TABLE_NAME IS NOT NULL ORDER BY k.TABLE_NAME,k.CONSTRAINT_NAME,k.ORDINAL_POSITION", schemaName)
		if queryErr != nil {
			return nil, queryErr
		}
		result := make([]map[string]interface{}, 0)
		positions := make(map[string]int)
		for _, record := range records {
			key := fmt.Sprint(record["table"]) + "\x00" + fmt.Sprint(record["name"])
			position, ok := positions[key]
			if !ok {
				position = len(result)
				positions[key] = position
				result = append(result, map[string]interface{}{"name": record["name"], "schema": record["schema"], "type": "FOREIGN KEY", "table": record["table"], "columns": []string{}, "reference_schema": record["reference_schema"], "reference_table": record["reference_table"], "reference_columns": []string{}, "on_delete": record["on_delete"], "on_update": record["on_update"]})
			}
			result[position]["columns"] = append(result[position]["columns"].([]string), fmt.Sprint(record["column_name"]))
			result[position]["reference_columns"] = append(result[position]["reference_columns"].([]string), fmt.Sprint(record["reference_column"]))
		}
		return result, nil
	default:
		return nil, errors.New("对象类型仅支持 view、routine、trigger 或 foreign_key")
	}
	args := []interface{}{schemaName}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		// LOCATE 按字面匹配，避免对象名称中的 % 和 _ 被当作 LIKE 通配符。
		source += " AND LOCATE(LOWER(?), LOWER(" + nameColumn + ")) > 0"
		args = append(args, keyword)
	}
	if !paginated {
		return m.records(ctx, "SELECT "+fields+definition+source+" ORDER BY "+order, args...)
	}
	// 统计和列表使用相同条件；列表不加载定义正文，完整定义由独立接口读取。
	var total int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*)"+source, args...).Scan(&total); err != nil {
		return nil, err
	}
	if total == 0 {
		pageNumber = 1
	} else if lastPage := (total-1)/int64(pageSize) + 1; int64(pageNumber) > lastPage {
		pageNumber = int(lastPage)
	}
	args = append(args, pageSize, int64(pageNumber-1)*int64(pageSize))
	list, err := m.records(ctx, "SELECT "+fields+source+" ORDER BY "+order+" LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, err
	}
	return page.New(&page.Query{Page: pageNumber, PageSize: pageSize}, total, list), nil
}

// ObjectDefinition 返回服务器的对象定义，保留存储程序及触发器正文。
func (m *Manager) ObjectDefinition(ctx context.Context, schemaName, kind, name string) (out string, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, target, err := m.target(schemaName, name)
	if err != nil {
		return "", err
	}
	statement := ""
	switch kind {
	case "view":
		statement = "SHOW CREATE VIEW " + target
	case "trigger":
		version, maria, versionErr := m.serverVersion(ctx)
		if versionErr != nil {
			return "", versionErr
		}
		if !maria && version < 50121 {
			objects, queryErr := m.Objects(ctx, schema, "trigger", "", 0, 0)
			if queryErr != nil {
				return "", queryErr
			}
			for _, record := range objects.([]map[string]interface{}) {
				if record["name"] != name {
					continue
				}
				_, table, _ := m.target(schema, fmt.Sprint(record["table"]))
				return "CREATE TRIGGER " + target + " " + fmt.Sprint(record["timing"]) + " " + fmt.Sprint(record["event"]) + " ON " + table + " FOR EACH ROW " + strings.TrimSuffix(strings.TrimSpace(fmt.Sprint(record["definition"])), ";") + ";", nil
			}
			return "", errors.New("触发器不存在或无权查看")
		}
		statement = "SHOW CREATE TRIGGER " + target
	case "routine", "procedure", "function":
		var routineType string
		if kind != "routine" {
			routineType = strings.ToUpper(kind)
		} else {
			err = m.db.QueryRowContext(ctx, "SELECT ROUTINE_TYPE FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = ? AND ROUTINE_NAME = ?", schema, name).Scan(&routineType)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("存储程序不存在或无权查看")
		}
		if err != nil {
			return "", err
		}
		if routineType != "FUNCTION" && routineType != "PROCEDURE" {
			return "", errors.New("存储程序类型不受支持")
		}
		statement = "SHOW CREATE " + routineType + " " + target
	case "foreign_key":
		objects, queryErr := m.Objects(ctx, schema, "foreign_key", "", 0, 0)
		if queryErr != nil {
			return "", queryErr
		}
		for _, record := range objects.([]map[string]interface{}) {
			if record["name"] != name {
				continue
			}
			key := ForeignKeyDefinition{Name: name, Columns: record["columns"].([]string), ReferenceSchema: fmt.Sprint(record["reference_schema"]), ReferenceTable: fmt.Sprint(record["reference_table"]), ReferenceColumns: record["reference_columns"].([]string), OnDelete: fmt.Sprint(record["on_delete"]), OnUpdate: fmt.Sprint(record["on_update"])}
			clause, buildErr := m.foreignKeySQL(schema, key)
			if buildErr != nil {
				return "", buildErr
			}
			_, table, _ := m.target(schema, fmt.Sprint(record["table"]))
			return "ALTER TABLE " + table + " ADD " + clause + ";", nil
		}
		return "", errors.New("外键不存在或无权查看")
	default:
		return "", errors.New("不支持的对象类型")
	}
	rows, err := m.db.QueryContext(ctx, statement)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return "", err
		}
		return "", errors.New("对象不存在或无权查看")
	}
	values := make([]interface{}, len(columns))
	pointers := make([]interface{}, len(columns))
	for index := range values {
		pointers[index] = &values[index]
	}
	if err = rows.Scan(pointers...); err != nil {
		return "", err
	}
	for index, column := range columns {
		label := strings.ToLower(column)
		if !strings.HasPrefix(label, "create ") && label != "sql original statement" {
			continue
		}
		switch value := values[index].(type) {
		case string:
			out = value
		case []byte:
			out = string(value)
		}
		break
	}
	if out == "" {
		return "", errors.New("当前账户没有查看对象完整定义的权限")
	}
	if len(out) > 1<<20 {
		return "", errors.New("对象定义超过 1 MB")
	}
	return strings.TrimSuffix(strings.TrimSpace(out), ";") + ";", nil
}

// objectSQL 验证对象编辑器只提交目标对象的一条 DDL，允许程序体内部的分号。
func (m *Manager) objectSQL(schemaName, kind, name, definition string) (string, string, error) {
	definition = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(definition), ";"))
	if definition == "" || len(definition) > 1<<20 || strings.ContainsRune(definition, 0) {
		return "", "", errors.New("对象定义不能为空且不能超过 1 MB")
	}
	identifier := "(?:`(?:[^`]|``)+`|[a-zA-Z_][a-zA-Z0-9_$]*)"
	account := "(?:'(?:''|[^'])*'|`(?:``|[^`])*`|[a-zA-Z0-9_%]+)"
	options := "(?:(?:ALGORITHM\\s*=\\s*(?:UNDEFINED|MERGE|TEMPTABLE)|DEFINER\\s*=\\s*(?:CURRENT_USER(?:\\(\\))?|" + account + "\\s*@\\s*" + account + ")|SQL\\s+SECURITY\\s+(?:DEFINER|INVOKER))\\s+)*"
	pattern := "(?is)^(CREATE(?:\\s+OR\\s+REPLACE)?|ALTER)\\s+" + options + "(VIEW|TRIGGER|PROCEDURE|FUNCTION)\\s+(" + identifier + ")(?:\\s*\\.\\s*(" + identifier + "))?(?:\\s|\\(|$)"
	match := regexp.MustCompile(pattern).FindStringSubmatch(definition)
	if match == nil {
		return "", "", errors.New("请输入完整的 CREATE 或 ALTER 对象定义")
	}
	objectType := strings.ToUpper(match[2])
	valid := (kind == "view" && objectType == "VIEW") || (kind == "trigger" && objectType == "TRIGGER") || (kind == "routine" && (objectType == "FUNCTION" || objectType == "PROCEDURE"))
	if !valid {
		return "", "", errors.New("对象定义与选择的对象类型不一致")
	}
	decode := func(value string) string {
		if strings.HasPrefix(value, "`") {
			return strings.ReplaceAll(value[1:len(value)-1], "``", "`")
		}
		return value
	}
	targetName, targetSchema := decode(match[3]), schemaName
	if match[4] != "" {
		targetSchema, targetName = decode(match[3]), decode(match[4])
	}
	if targetSchema == "" {
		targetSchema = m.database
	}
	if targetSchema != m.database || targetName != name {
		return "", "", errors.New("对象定义必须使用当前选择的数据库和对象名称")
	}
	// 程序体内部允许分号，顶层分号只能是最后一个语句终止符。
	quote, depth, blocks := byte(0), 0, 0
	previous := ""
	word := strings.Builder{}
	flush := func(next byte) {
		token := strings.ToUpper(word.String())
		word.Reset()
		if token == "" {
			return
		}
		qualifier := previous == "END" && (token == "IF" || token == "CASE" || token == "LOOP" || token == "WHILE" || token == "REPEAT")
		if !qualifier {
			switch token {
			case "BEGIN", "CASE":
				blocks++
			case "IF":
				if blocks > 0 && next != '(' {
					blocks++
				}
			case "LOOP", "WHILE", "REPEAT":
				if blocks > 0 {
					blocks++
				}
			case "END":
				if blocks > 0 {
					blocks--
				}
			}
		}
		previous = token
	}
	for index := 0; index < len(definition); index++ {
		ch := definition[index]
		if quote != 0 {
			if ch == '\\' {
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
		next := ch
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
			remaining := strings.TrimLeft(definition[index:], " \n\r\t")
			if remaining != "" {
				next = remaining[0]
			}
		}
		flush(next)
		switch ch {
		case '\'', '"', '`':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return "", "", errors.New("对象定义括号不匹配")
			}
		case ';':
			if blocks == 0 && depth == 0 {
				return "", "", errors.New("只允许提交一条对象定义")
			}
		case '#':
			if end := strings.IndexByte(definition[index:], '\n'); end >= 0 {
				index += end
			} else {
				index = len(definition)
			}
		case '-', '/':
			if index+1 < len(definition) && ch == '-' && definition[index+1] == '-' {
				if end := strings.IndexByte(definition[index:], '\n'); end >= 0 {
					index += end
				} else {
					index = len(definition)
				}
			} else if index+1 < len(definition) && ch == '/' && definition[index+1] == '*' {
				if strings.HasPrefix(definition[index:], "/*!") {
					return "", "", errors.New("对象定义不支持可执行版本注释")
				}
				end := strings.Index(definition[index+2:], "*/")
				if end < 0 {
					return "", "", errors.New("对象定义注释未结束")
				}
				index += end + 3
			}
		}
	}
	flush(0)
	if quote != 0 || depth != 0 || blocks != 0 {
		return "", "", errors.New("对象定义的引号、括号或 BEGIN / END 不匹配")
	}

	return definition, objectType, nil
}

// SaveObject 保存真实对象定义；存储程序及触发器替换失败时恢复原定义与授权。
func (m *Manager) SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, _, err := m.target(schemaName, name)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(schema) {
	case "", "mysql", "information_schema", "performance_schema", "sys":
		return nil, errors.New("不能修改系统数据库中的对象")
	}
	statement, objectType, err := m.objectSQL(schema, kind, name, definition)
	if err != nil {
		return nil, err
	}
	if replace && strings.HasPrefix(strings.ToUpper(statement), "CREATE") && !strings.HasPrefix(strings.ToUpper(statement), "CREATE OR REPLACE") {
		if objectType == "VIEW" {
			statement = "CREATE OR REPLACE" + statement[len("CREATE"):]
		}
	}
	if objectType != "VIEW" && strings.HasPrefix(strings.ToUpper(statement), "CREATE OR REPLACE") {
		return nil, errors.New("MySQL 存储程序与触发器不支持 CREATE OR REPLACE")
	}
	if !replace || objectType == "VIEW" || strings.HasPrefix(strings.ToUpper(statement), "ALTER") {
		return m.executeMaintenance(ctx, statement)
	}
	// MySQL 没有可替换程序体的 ALTER；确认修改时保存原定义，失败恢复旧对象。
	// 此过程涉及隐式提交，不能宣称与事务内 DDL 一样原子。
	connection, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	lockName := fmt.Sprintf("panel.object.%x", sha256.Sum256([]byte(schema+"\x00"+kind+"\x00"+name)))[:64]
	var locked sql.NullInt64
	if err = connection.QueryRowContext(ctx, "SELECT GET_LOCK(?, 5)", lockName).Scan(&locked); err != nil {
		return nil, err
	}
	if !locked.Valid || locked.Int64 != 1 {
		return nil, errors.New("当前对象正在修改，请稍后重试")
	}
	defer func() {
		release, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		_, _ = connection.ExecContext(release, "DO RELEASE_LOCK(?)", lockName)
	}()
	backup, err := m.ObjectDefinition(ctx, schema, kind, name)
	if err != nil {
		return nil, err
	}
	// 程序的专有授权会随 DROP 消失，先保存后恢复；触发器复用表权限。
	grants := make([]string, 0)
	if kind == "routine" {
		var currentType string
		if err = m.db.QueryRowContext(ctx, "SELECT ROUTINE_TYPE FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = ? AND ROUTINE_NAME = ?", schema, name).Scan(&currentType); err != nil {
			return nil, err
		}
		if currentType != objectType {
			return nil, errors.New("存储程序修改不能改变 FUNCTION / PROCEDURE 类型")
		}
		records, queryErr := m.records(ctx, "SELECT Host AS host, User AS user, Proc_priv AS privileges FROM mysql.procs_priv WHERE Db = ? AND Routine_name = ? AND Routine_type = ?", schema, name, objectType)
		if queryErr != nil {
			return nil, fmt.Errorf("无法备份程序授权，原对象未修改：%w", queryErr)
		}
		for _, record := range records {
			permissions := make([]string, 0)
			grantOption := false
			for _, permission := range strings.Split(fmt.Sprint(record["privileges"]), ",") {
				switch strings.ToUpper(strings.TrimSpace(permission)) {
				case "EXECUTE":
					permissions = append(permissions, "EXECUTE")
				case "ALTER ROUTINE":
					permissions = append(permissions, "ALTER ROUTINE")
				case "GRANT":
					grantOption = true
				}
			}
			if len(permissions) == 0 {
				continue
			}
			user, literalErr := m.schemaLiteral(ctx, fmt.Sprint(record["user"]))
			if literalErr != nil {
				return nil, literalErr
			}
			host, literalErr := m.schemaLiteral(ctx, fmt.Sprint(record["host"]))
			if literalErr != nil {
				return nil, literalErr
			}
			_, target, _ := m.target(schema, name)
			grant := "GRANT " + strings.Join(permissions, ", ") + " ON " + objectType + " " + target + " TO " + user + "@" + host
			if grantOption {
				grant += " WITH GRANT OPTION"
			}
			grants = append(grants, grant)
		}
	}
	if kind == "trigger" {
		// 老版本只允许同一事件的一个触发器；新版本保留原有多触发器的先后顺序。
		var ordered int
		if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'information_schema' AND TABLE_NAME = 'TRIGGERS' AND COLUMN_NAME = 'ACTION_ORDER'").Scan(&ordered); err != nil {
			return nil, err
		}
		if ordered > 0 {
			neighbors, queryErr := m.records(ctx, "SELECT t.TRIGGER_NAME AS name, o.ACTION_TIMING AS timing, o.EVENT_MANIPULATION AS event, o.EVENT_OBJECT_TABLE AS target_table, CASE WHEN t.ACTION_ORDER < o.ACTION_ORDER THEN 'FOLLOWS' ELSE 'PRECEDES' END AS ordering FROM information_schema.TRIGGERS o JOIN information_schema.TRIGGERS t ON t.TRIGGER_SCHEMA = o.TRIGGER_SCHEMA AND t.EVENT_OBJECT_TABLE = o.EVENT_OBJECT_TABLE AND t.ACTION_TIMING = o.ACTION_TIMING AND t.EVENT_MANIPULATION = o.EVENT_MANIPULATION AND t.TRIGGER_NAME <> o.TRIGGER_NAME WHERE o.TRIGGER_SCHEMA = ? AND o.TRIGGER_NAME = ? ORDER BY (t.ACTION_ORDER < o.ACTION_ORDER) DESC, ABS(t.ACTION_ORDER-o.ACTION_ORDER) LIMIT 1", schema, name)
			if queryErr != nil {
				return nil, queryErr
			}
			if len(neighbors) > 0 {
				neighbor := neighbors[0]
				identifier := "(?:`(?:[^`]|``)+`|[a-zA-Z_][a-zA-Z0-9_$]*)"
				header := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(fmt.Sprint(neighbor["timing"])) + `\s+` + regexp.QuoteMeta(fmt.Sprint(neighbor["event"])) + `\s+ON\s+(` + identifier + `)(?:\s*\.\s*(` + identifier + `))?\s+`).FindStringSubmatch(statement)
				if header != nil {
					decode := func(value string) string {
						if strings.HasPrefix(value, "`") {
							return strings.ReplaceAll(value[1:len(value)-1], "``", "`")
						}
						return value
					}
					targetSchema, targetTable := schema, decode(header[1])
					if header[2] != "" {
						targetSchema, targetTable = decode(header[1]), decode(header[2])
					}
					if targetSchema == schema && targetTable == fmt.Sprint(neighbor["target_table"]) {
						adjacent, _ := m.quote(fmt.Sprint(neighbor["name"]))
						location := regexp.MustCompile(`(?i)\bFOR\s+EACH\s+ROW\b`).FindStringIndex(statement)
						if location == nil {
							return nil, errors.New("触发器定义缺少 FOR EACH ROW")
						}
						ordering := " " + fmt.Sprint(neighbor["ordering"]) + " " + adjacent
						hasOrder := regexp.MustCompile(`(?i)\b(?:FOLLOWS|PRECEDES)\b`)
						if !hasOrder.MatchString(statement) {
							statement = statement[:location[1]] + ordering + statement[location[1]:]
						}
						if !hasOrder.MatchString(backup) {
							if original := regexp.MustCompile(`(?i)\bFOR\s+EACH\s+ROW\b`).FindStringIndex(backup); original != nil {
								backup = backup[:original[1]] + ordering + backup[original[1]:]
							}
						}
					}
				}
			}
		}
	}
	_, target, _ := m.target(schema, name)
	drop := "DROP " + objectType + " " + target
	started := time.Now()
	if _, err = connection.ExecContext(ctx, drop); err != nil {
		return nil, err
	}
	restore := func(cause error) error {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_, _ = connection.ExecContext(cleanup, "DROP "+objectType+" IF EXISTS "+target)
		if _, restoreErr := connection.ExecContext(cleanup, backup); restoreErr != nil {
			return fmt.Errorf("新定义保存失败：%v；恢复旧定义也失败：%v；原始定义：\n%s", cause, restoreErr, backup)
		}
		for _, grant := range grants {
			if _, restoreErr := connection.ExecContext(cleanup, grant); restoreErr != nil {
				return fmt.Errorf("新定义保存失败：%v；原定义已恢复，但授权恢复失败：%v", cause, restoreErr)
			}
		}
		return fmt.Errorf("新定义保存失败，已恢复原对象及授权：%w", cause)
	}
	result, executeErr := connection.ExecContext(ctx, statement)
	if executeErr != nil {
		return nil, restore(executeErr)
	}
	for _, grant := range grants {
		if _, executeErr = connection.ExecContext(ctx, grant); executeErr != nil {
			return nil, restore(executeErr)
		}
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, Affected: affected, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// DropObject 删除真实视图、存储程序或触发器，外键仍定位到所属表后删除。
func (m *Manager) DropObject(ctx context.Context, schemaName, kind, name string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schema, target, err := m.target(schemaName, name)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(schema) {
	case "", "mysql", "information_schema", "performance_schema", "sys":
		return nil, errors.New("不能删除系统数据库中的对象")
	}
	switch kind {
	case "view":
		return m.executeMaintenance(ctx, "DROP VIEW "+target)
	case "trigger":
		return m.executeMaintenance(ctx, "DROP TRIGGER "+target)
	case "routine":
		var routineType string
		err = m.db.QueryRowContext(ctx, "SELECT ROUTINE_TYPE FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = ? AND ROUTINE_NAME = ?", schema, name).Scan(&routineType)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("存储程序不存在")
		}
		if err != nil {
			return nil, err
		}
		if routineType != "FUNCTION" && routineType != "PROCEDURE" {
			return nil, errors.New("存储程序类型不受支持")
		}
		return m.executeMaintenance(ctx, "DROP "+routineType+" "+target)
	case "foreign_key":
		objects, queryErr := m.Objects(ctx, schema, "foreign_key", "", 0, 0)
		if queryErr != nil {
			return nil, queryErr
		}
		for _, record := range objects.([]map[string]interface{}) {
			if record["name"] == name {
				return m.DropForeignKey(ctx, schema, fmt.Sprint(record["table"]), name)
			}
		}
		return nil, errors.New("外键不存在")
	default:
		return nil, errors.New("不支持的对象类型")
	}
}
