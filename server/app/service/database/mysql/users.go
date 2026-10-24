package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"server/app/util/page"
	"strings"
	"time"
	"unicode/utf8"
)

// Users 分页列出账号及其在所有数据库的授权，角色授权单独返回，不将对象权限误认为整库权限。
func (m *Manager) Users(ctx context.Context, keyword string, pageNumber, pageSize int) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if pageNumber < 1 || pageNumber > 1000000 || pageSize < 1 || pageSize > 1000 {
		return nil, errors.New("页码应为 1 至 1000000，每页数量应为 1 至 1000")
	}
	// 内部账户不参与列表、搜索和分页计数。
	source := " FROM mysql.user WHERE User NOT IN ('mysql.infoschema', 'mysql.session', 'mysql.sys')"
	args := make([]interface{}, 0, 4)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		// LOCATE 按字面匹配，账号及来源中的 % 和 _ 不作为通配符。
		source += " AND (LOCATE(LOWER(?), LOWER(User)) > 0 OR LOCATE(LOWER(?), LOWER(Host)) > 0)"
		args = append(args, keyword, keyword)
	}
	var total int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*)"+source, args...).Scan(&total); err != nil {
		return nil, err
	}
	if total == 0 {
		return page.New(&page.Query{Page: 1, PageSize: pageSize}, total, []*User{}), nil
	}
	if lastPage := (total-1)/int64(pageSize) + 1; int64(pageNumber) > lastPage {
		pageNumber = int(lastPage)
	}

	// 不同版本的权限列及角色表不同，只选择已存在的权限列，不读取密码和认证插件字段。
	metadata, err := m.records(ctx, "SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'mysql' AND TABLE_NAME IN ('user', 'db', 'tables_priv', 'columns_priv', 'procs_priv', 'global_grants', 'role_edges', 'roles_mapping') ORDER BY TABLE_NAME, ORDINAL_POSITION")
	if err != nil {
		return nil, err
	}
	columns := make(map[string][]string)
	available := make(map[string]bool)
	for _, column := range metadata {
		table, name := strings.ToLower(fmt.Sprint(column["table_name"])), strings.ToLower(fmt.Sprint(column["column_name"]))
		available[table] = true
		available[table+"."+name] = true
		if strings.HasSuffix(name, "_priv") {
			columns[table] = append(columns[table], name)
		}
	}
	if len(columns["user"]) == 0 || len(columns["db"]) == 0 {
		return nil, errors.New("无法读取账号权限字段，请检查当前连接的授权")
	}
	privilegeName := func(name string) string {
		name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), "_priv")
		if label, ok := map[string]string{"grant": "GRANT OPTION", "show_db": "SHOW DATABASES", "create_tmp_table": "CREATE TEMPORARY TABLES", "repl_slave": "REPLICATION SLAVE", "repl_client": "REPLICATION CLIENT"}[name]; ok {
			return label
		}
		return strings.ToUpper(strings.ReplaceAll(name, "_", " "))
	}
	privileges := func(record map[string]interface{}, table string) []string {
		result := make([]string, 0)
		for _, column := range columns[table] {
			if strings.EqualFold(fmt.Sprint(record[column]), "Y") {
				result = append(result, privilegeName(column))
			}
		}
		return result
	}
	selectFields := "User AS name, Host AS host, BINARY CONCAT(User, '@', Host) = BINARY CURRENT_USER() AS current_account, " + strings.Join(columns["user"], ", ")
	if available["user.user_attributes"] {
		// 仅取部分撤权信息；User_attributes 的其他内容可能包含认证信息。
		selectFields += ", JSON_EXTRACT(User_attributes, '$.Restrictions') AS restrictions"
	}
	args = append(args, pageSize, int64(pageNumber-1)*int64(pageSize))
	records, err := m.records(ctx, "SELECT "+selectFields+source+" ORDER BY User, Host LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, err
	}
	list := make([]*User, 0, len(records))
	users := make(map[string]*User, len(records))
	accounts := make([]string, 0, len(records))
	grantArgs := make([]interface{}, 0, len(records)*2)
	for _, record := range records {
		name, host := fmt.Sprint(record["name"]), fmt.Sprint(record["host"])
		internal := name == "mariadb.sys" || strings.HasPrefix(name, "mysql.")
		current := m.truth(record["current_account"])
		user := &User{Name: name, Host: host, GlobalPrivileges: privileges(record, "user"), Grants: []UserGrant{}, GrantedRoles: []string{}, Restrictions: []UserRestriction{},
			Current: current, PasswordEditable: !internal && name != "",
			Protected: strings.EqualFold(fmt.Sprint(record["super_priv"]), "Y") || current || name == "root" || internal}
		if value := record["restrictions"]; value != nil && fmt.Sprint(value) != "null" {
			if err = json.Unmarshal([]byte(fmt.Sprint(value)), &user.Restrictions); err != nil {
				return nil, errors.New("读取账号的数据库权限限制失败")
			}
		}
		list = append(list, user)
		users[name+"\x00"+strings.ToLower(host)] = user
		accounts = append(accounts, "(?, ?)")
		grantArgs = append(grantArgs, name, host)
	}
	if len(list) == 0 {
		return page.New(&page.Query{Page: pageNumber, PageSize: pageSize}, total, list), nil
	}
	condition := " WHERE (User, Host) IN (" + strings.Join(accounts, ", ") + ")"
	var variable, partial string
	err = m.db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'partial_revokes'").Scan(&variable, &partial)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	literalDatabase := strings.EqualFold(partial, "ON") || partial == "1"
	grants, err := m.records(ctx, "SELECT User AS name, Host AS host, Db AS database_name, "+strings.Join(columns["db"], ", ")+" FROM mysql.db"+condition+" ORDER BY User, Host, Db", grantArgs...)
	if err != nil {
		return nil, err
	}
	for _, record := range grants {
		user := users[fmt.Sprint(record["name"])+"\x00"+strings.ToLower(fmt.Sprint(record["host"]))]
		grant := UserGrant{Database: fmt.Sprint(record["database_name"]), Scope: "database", Privileges: privileges(record, "db")}
		if !literalDatabase {
			// mysql.db 的库名采用授权模式。转义字符还原为真实名称，未转义通配符保留标识。
			var name strings.Builder
			for index := 0; index < len(grant.Database); index++ {
				value := grant.Database[index]
				if value == '\\' && index+1 < len(grant.Database) {
					index++
					value = grant.Database[index]
				} else if value == '%' || value == '_' {
					grant.Pattern = true
				}
				name.WriteByte(value)
			}
			grant.Database = name.String()
		}
		if len(grant.Privileges) > 0 {
			user.Grants = append(user.Grants, grant)
		}
	}
	// 各对象授权分别显示；表中的 Column_priv 只是汇总标记，具体列权限从 columns_priv 读取。
	for _, table := range []string{"tables_priv", "columns_priv", "procs_priv"} {
		if !available[table] {
			continue
		}
		fields, scope := "Table_name AS object_name, Table_priv AS privileges", "table"
		if table == "columns_priv" {
			fields, scope = "CONCAT(Table_name, '.', Column_name) AS object_name, Column_priv AS privileges", "column"
		} else if table == "procs_priv" {
			fields, scope = "CONCAT(Routine_type, ' ', Routine_name) AS object_name, Proc_priv AS privileges", "routine"
		}
		grants, err = m.records(ctx, "SELECT User AS name, Host AS host, Db AS database_name, "+fields+" FROM mysql."+table+condition+" ORDER BY User, Host, Db, object_name", grantArgs...)
		if err != nil {
			return nil, err
		}
		for _, record := range grants {
			grant := UserGrant{Database: fmt.Sprint(record["database_name"]), Scope: scope, Object: fmt.Sprint(record["object_name"]), Privileges: []string{}}
			for _, privilege := range strings.Split(fmt.Sprint(record["privileges"]), ",") {
				if privilege = strings.TrimSpace(privilege); privilege != "" {
					grant.Privileges = append(grant.Privileges, privilegeName(privilege))
				}
			}
			if len(grant.Privileges) > 0 {
				user := users[fmt.Sprint(record["name"])+"\x00"+strings.ToLower(fmt.Sprint(record["host"]))]
				user.Grants = append(user.Grants, grant)
			}
		}
	}
	if available["global_grants"] {
		grants, err = m.records(ctx, "SELECT USER AS name, HOST AS host, PRIV AS privilege_name FROM mysql.global_grants"+condition+" ORDER BY USER, HOST, PRIV", grantArgs...)
		if err != nil {
			return nil, err
		}
		for _, record := range grants {
			user := users[fmt.Sprint(record["name"])+"\x00"+strings.ToLower(fmt.Sprint(record["host"]))]
			user.GlobalPrivileges = append(user.GlobalPrivileges, fmt.Sprint(record["privilege_name"]))
		}
	}
	for _, table := range []string{"role_edges", "roles_mapping"} {
		if !available[table] {
			continue
		}
		query := "SELECT User AS name, Host AS host, `Role` AS role_name FROM mysql.roles_mapping" + condition + " ORDER BY User, Host, `Role`"
		if table == "role_edges" {
			query = "SELECT TO_USER AS name, TO_HOST AS host, CONCAT(FROM_USER, '@', FROM_HOST) AS role_name FROM mysql.role_edges WHERE (TO_USER, TO_HOST) IN (" + strings.Join(accounts, ", ") + ") ORDER BY TO_USER, TO_HOST, FROM_USER, FROM_HOST"
		}
		grants, err = m.records(ctx, query, grantArgs...)
		if err != nil {
			return nil, err
		}
		for _, record := range grants {
			user := users[fmt.Sprint(record["name"])+"\x00"+strings.ToLower(fmt.Sprint(record["host"]))]
			user.GrantedRoles = append(user.GrantedRoles, fmt.Sprint(record["role_name"]))
		}
	}
	if available["role_edges"] {
		var mandatory string
		err = m.db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'mandatory_roles'").Scan(&variable, &mandatory)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if mandatory != "" {
			for _, user := range list {
				user.GrantedRoles = append(user.GrantedRoles, "强制角色："+mandatory)
			}
		}
	}
	for _, user := range list {
		user.Roles = strings.Join(user.GlobalPrivileges, ", ")
	}
	return page.New(&page.Query{Page: pageNumber, PageSize: pageSize}, total, list), nil
}

// CreateUser 创建独立 MySQL 账号，不自动授予数据库权限。
func (m *Manager) CreateUser(ctx context.Context, name, host, password string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	account, exists, err := m.account(ctx, name, host, false)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("该账号与来源地址已存在")
	}
	if password == "" || len(password) > 1024 || strings.ContainsRune(password, 0) {
		return nil, errors.New("密码应为 1 至 1024 字节，且不能包含空字符")
	}

	var maximum int
	if err = m.db.QueryRowContext(ctx, "SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='mysql' AND TABLE_NAME='user' AND COLUMN_NAME='User'").Scan(&maximum); err != nil {
		return nil, err
	}
	if maximum < 1 || utf8.RuneCountInString(name) > maximum {
		return nil, fmt.Errorf("当前数据库版本的账号名称最多 %d 个字符", maximum)
	}
	conn, close, err := m.accountConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	result, err := m.accountExec(ctx, conn, "CREATE USER "+account+" IDENTIFIED BY "+m.accountLiteral(password))
	if err != nil {
		return nil, errors.New("账号或密码设置失败，请检查账号权限、认证方式和密码策略")
	}
	return result, nil
}

// ChangePassword 修改账号密码，返回是否为当前连接账户，供调用方同步连接凭据。
func (m *Manager) ChangePassword(ctx context.Context, name, host, password string) (out *Result, current bool, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, false, err
	}
	if host == "" {
		host = "%"
	}
	account, exists, err := m.account(ctx, name, host, true)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, errors.New("账号不存在，请刷新列表")
	}
	if password == "" || len(password) > 1024 || strings.ContainsRune(password, 0) {
		return nil, false, errors.New("密码应为 1 至 1024 字节，且不能包含空字符")
	}
	version, maria, err := m.serverVersion(ctx)
	if err != nil {
		return nil, false, err
	}
	conn, close, err := m.accountConnection(ctx)
	if err != nil {
		return nil, false, err
	}
	defer close()
	statement := ""
	var connectedAccount string
	if err = conn.QueryRowContext(ctx, "SELECT CURRENT_USER()").Scan(&connectedAccount); err != nil {
		return nil, false, err
	}
	if index := strings.LastIndexByte(connectedAccount, '@'); index >= 0 {
		current = connectedAccount[:index] == name && strings.EqualFold(connectedAccount[index+1:], host)
	}
	if maria || version < 50706 {
		// MariaDB 的 SET PASSWORD 保留 root 的多种认证方式（如 unix_socket）。
		statement = "SET PASSWORD FOR " + account + " = PASSWORD(" + m.accountLiteral(password) + ")"
	} else {
		var plugin string
		if err = conn.QueryRowContext(ctx, "SELECT plugin FROM mysql.user WHERE BINARY User = ? AND Host = ?", name, host).Scan(&plugin); err != nil {
			return nil, false, err
		}
		if plugin == "" || plugin == "auth_socket" {
			return nil, false, errors.New("该账户未使用密码认证，无法直接修改密码")
		}
		// 显式保留原认证插件，避免改密时切换到服务器默认插件。
		statement = "ALTER USER " + account + " IDENTIFIED WITH " + m.accountLiteral(plugin) + " BY " + m.accountLiteral(password)
	}
	result, err := m.accountExec(ctx, conn, statement)
	if err != nil {
		return nil, false, errors.New("账号或密码设置失败，请检查账号权限、认证方式和密码策略")
	}
	return result, current, nil
}

// ChangeAccess 修改账号的连接来源，保留原账号的密码和已有授权。
func (m *Manager) ChangeAccess(ctx context.Context, name, host, newHost string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if host == "" {
		host = "%"
	}
	newHost = strings.TrimSpace(newHost)
	if newHost == "" {
		return nil, errors.New("请选择新的连接来源")
	}
	// root 可调整来源，删除及权限变更仍由各自操作保留保护。
	account, exists, err := m.account(ctx, name, host, name == "root")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	if strings.EqualFold(host, newHost) {
		return &Result{Columns: []string{}, Rows: [][]interface{}{}}, nil
	}
	target, exists, err := m.account(ctx, name, newHost, name == "root")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("该账号在新的连接来源下已存在，请选择其他来源")
	}
	conn, close, err := m.accountConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	// 使用原生 RENAME USER，避免删建账号造成密码、权限或认证插件丢失。
	return m.accountExec(ctx, conn, "RENAME USER "+account+" TO "+target)
}

// Permissions 读取账号可编辑的直接整库权限；未指定账号时仅返回当前版本支持的权限。
func (m *Manager) Permissions(ctx context.Context, name, host string) (out *UserPermissions, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if host == "" {
		host = "%"
	}
	if name != "" {
		_, exists, err := m.account(ctx, name, host, false)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("账号不存在，请刷新列表")
		}
	}
	out, _, err = m.accountPermissions(ctx, name, host)
	return out, err
}

// UpdatePermissions 按提交的最终状态计算差量，只修改提交数据库的直接整库权限。
func (m *Manager) UpdatePermissions(ctx context.Context, name, host string, grants []DatabasePrivileges) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	if host == "" {
		host = "%"
	}
	account, exists, err := m.account(ctx, name, host, false)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	if len(grants) > 1000 {
		return nil, errors.New("一次最多调整 1000 个数据库的权限")
	}
	permissions, literalDatabase, err := m.accountPermissions(ctx, name, host)
	if err != nil {
		return nil, err
	}
	var lowerCaseNames int
	if err = m.db.QueryRowContext(ctx, "SELECT @@lower_case_table_names").Scan(&lowerCaseNames); err != nil {
		return nil, err
	}
	key := func(database string) string {
		if lowerCaseNames != 0 {
			return strings.ToLower(database)
		}
		return database
	}
	current := make(map[string]map[string]bool, len(permissions.Grants))
	for _, grant := range permissions.Grants {
		database := key(grant.Database)
		if _, exists := current[database]; exists {
			return nil, errors.New("数据库授权记录存在重名，请先检查账号的原始授权")
		}
		current[database] = make(map[string]bool, len(grant.Privileges))
		for _, privilege := range grant.Privileges {
			current[database][privilege] = true
		}
	}
	// GRANT 会解除同库、同权限的部分撤权，先检测，避免修改全局授权的限制。
	var restrictions []UserRestriction
	if literalDatabase {
		var content sql.NullString
		if err = m.db.QueryRowContext(ctx, "SELECT JSON_EXTRACT(User_attributes, '$.Restrictions') FROM mysql.user WHERE BINARY User = ? AND Host = ?", name, host).Scan(&content); err != nil {
			return nil, err
		}
		if content.Valid && content.String != "" && content.String != "null" {
			if err = json.Unmarshal([]byte(content.String), &restrictions); err != nil {
				return nil, errors.New("读取账号的数据库权限限制失败")
			}
		}
	}
	allowed := make(map[string]bool, len(permissions.Privileges))
	for _, privilege := range permissions.Privileges {
		allowed[privilege] = true
	}
	seen := make(map[string]bool, len(grants))
	statements := make([]string, 0, len(grants)*2)
	for _, grant := range grants {
		database := grant.Database
		if _, err = m.quote(database); err != nil {
			return nil, errors.New("数据库名称不合法")
		}
		switch strings.ToLower(database) {
		case "mysql", "information_schema", "performance_schema", "sys":
			return nil, errors.New("不能调整系统数据库的权限")
		}
		databaseKey := key(database)
		if seen[databaseKey] {
			return nil, errors.New("同一个数据库不能重复提交")
		}
		seen[databaseKey] = true
		desired := make(map[string]bool, len(grant.Privileges))
		for _, privilege := range grant.Privileges {
			privilege = strings.ToUpper(strings.TrimSpace(privilege))
			if !allowed[privilege] {
				return nil, errors.New("包含当前版本不支持管理的数据库权限")
			}
			desired[privilege] = true
		}
		add, remove := make([]string, 0), make([]string, 0)
		for _, privilege := range permissions.Privileges {
			if desired[privilege] && !current[databaseKey][privilege] {
				add = append(add, privilege)
			} else if !desired[privilege] && current[databaseKey][privilege] {
				remove = append(remove, privilege)
			}
		}
		if len(add) > 0 {
			condition := "BINARY SCHEMA_NAME = ?"
			if lowerCaseNames != 0 {
				condition = "LOWER(SCHEMA_NAME) = LOWER(?)"
			}
			if err = m.db.QueryRowContext(ctx, "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE "+condition, database).Scan(&database); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, fmt.Errorf("数据库 %s 不存在或当前连接无权访问", grant.Database)
				}
				return nil, err
			}
			for _, restriction := range restrictions {
				if key(restriction.Database) != databaseKey {
					continue
				}
				for _, privilege := range restriction.Privileges {
					for _, added := range add {
						if strings.EqualFold(privilege, added) {
							return nil, fmt.Errorf("数据库 %s 的 %s 存在全局权限限制，请先单独处理该限制", database, added)
						}
					}
				}
			}
		}
		if !literalDatabase {
			database = strings.NewReplacer(`\`, `\\`, "_", `\_`, "%", `\%`).Replace(database)
		}
		target, err := m.quote(database)
		if err != nil {
			return nil, err
		}
		if len(add) > 0 {
			statements = append(statements, "GRANT "+strings.Join(add, ", ")+" ON "+target+".* TO "+account)
		}
		if len(remove) > 0 {
			statements = append(statements, "REVOKE "+strings.Join(remove, ", ")+" ON "+target+".* FROM "+account)
		}
	}
	result := &Result{Columns: []string{}, Rows: [][]interface{}{}}
	if len(statements) > 0 {
		conn, close, err := m.accountConnection(ctx)
		if err != nil {
			return nil, err
		}
		defer close()
		// 账号语句会隐式提交。重试时重新读取差量，不重复撤回已删除的授权。
		for _, statement := range statements {
			changed, err := m.accountExec(ctx, conn, statement)
			if err != nil {
				return nil, fmt.Errorf("权限更新未完成，部分更改可能已生效，请重试：%w", err)
			}
			result.Affected += changed.Affected
		}
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// accountPermissions 只将精确库名和已知权限开放编辑，其他授权保持原样。
func (m *Manager) accountPermissions(ctx context.Context, name, host string) (*UserPermissions, bool, error) {
	known := map[string]string{
		"select_priv": "SELECT", "insert_priv": "INSERT", "update_priv": "UPDATE", "delete_priv": "DELETE",
		"create_priv": "CREATE", "drop_priv": "DROP", "references_priv": "REFERENCES", "index_priv": "INDEX", "alter_priv": "ALTER",
		"create_tmp_table_priv": "CREATE TEMPORARY TABLES", "lock_tables_priv": "LOCK TABLES",
		"create_view_priv": "CREATE VIEW", "show_view_priv": "SHOW VIEW", "create_routine_priv": "CREATE ROUTINE",
		"alter_routine_priv": "ALTER ROUTINE", "execute_priv": "EXECUTE", "event_priv": "EVENT", "trigger_priv": "TRIGGER", "delete_history_priv": "DELETE HISTORY",
	}
	metadata, err := m.records(ctx, "SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'mysql' AND TABLE_NAME = 'db' ORDER BY ORDINAL_POSITION")
	if err != nil {
		return nil, false, err
	}
	result := &UserPermissions{Privileges: []string{}, Grants: []DatabasePrivileges{}}
	columns, fields := make([]string, 0), make([]string, 0)
	for _, record := range metadata {
		column := strings.ToLower(fmt.Sprint(record["column_name"]))
		if !strings.HasSuffix(column, "_priv") {
			continue
		}
		field, err := m.quote(column)
		if err != nil {
			return nil, false, err
		}
		columns, fields = append(columns, column), append(fields, field)
		if privilege, ok := known[column]; ok {
			result.Privileges = append(result.Privileges, privilege)
		}
	}
	if len(result.Privileges) == 0 {
		return nil, false, errors.New("无法读取当前版本支持的数据库权限")
	}
	// 创建账号时只需要权限选项，不读取同名匿名账号的授权。
	if name == "" {
		return result, false, nil
	}
	var variable, partial string
	err = m.db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'partial_revokes'").Scan(&variable, &partial)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	literalDatabase := strings.EqualFold(partial, "ON") || partial == "1"
	grants, err := m.records(ctx, "SELECT Db AS database_name, "+strings.Join(fields, ", ")+" FROM mysql.db WHERE BINARY User = ? AND Host = ? ORDER BY Db", name, host)
	if err != nil {
		return nil, false, err
	}
	for _, record := range grants {
		raw := fmt.Sprint(record["database_name"])
		database, pattern := raw, false
		if !literalDatabase {
			var decoded strings.Builder
			for index := 0; index < len(raw); index++ {
				value := raw[index]
				if value == '\\' && index+1 < len(raw) {
					index++
					value = raw[index]
				} else if value == '%' || value == '_' {
					pattern = true
				}
				decoded.WriteByte(value)
			}
			database = decoded.String()
			// 非标准转义不改写成另一个 mysql.db 记录。
			pattern = pattern || strings.NewReplacer(`\`, `\\`, "_", `\_`, "%", `\%`).Replace(database) != raw
		}
		switch strings.ToLower(database) {
		case "mysql", "information_schema", "performance_schema", "sys":
			pattern = true
		}
		grant := DatabasePrivileges{Database: database, Privileges: []string{}}
		for _, column := range columns {
			if !strings.EqualFold(fmt.Sprint(record[column]), "Y") {
				continue
			}
			if privilege, ok := known[column]; ok && !pattern {
				grant.Privileges = append(grant.Privileges, privilege)
			} else {
				result.Retained = true
			}
		}
		if !pattern {
			result.Grants = append(result.Grants, grant)
		}
	}
	return result, literalDatabase, nil
}

// DeleteUser 删除业务账号，保护当前连接与系统管理员。
func (m *Manager) DeleteUser(ctx context.Context, name, host string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	account, exists, err := m.account(ctx, name, host, false)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	conn, close, err := m.accountConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	return m.accountExec(ctx, conn, "DROP USER "+account)
}

// account 校验目标账号并读取保护状态，调用方决定创建或修改行为。
func (m *Manager) account(ctx context.Context, name, host string, allowProtected bool) (string, bool, error) {
	if host == "" {
		host = "%"
	}
	if name == "" || utf8.RuneCountInString(name) > 128 || strings.ContainsRune(name, 0) || len(host) > 255 || strings.ContainsRune(host, 0) {
		return "", false, errors.New("账号名称或来源地址不合法")
	}
	if name == "root" && !allowProtected || name == "mariadb.sys" || strings.HasPrefix(name, "mysql.") {
		return "", false, errors.New("不能修改系统账号")
	}
	var super string
	var current bool
	err := m.db.QueryRowContext(ctx, "SELECT Super_priv, BINARY CONCAT(User, '@', Host) = BINARY CURRENT_USER() FROM mysql.user WHERE BINARY User = ? AND Host = ?", name, host).Scan(&super, &current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	exists := err == nil
	if exists && !allowProtected && (strings.EqualFold(super, "Y") || current) {
		return "", false, errors.New("当前连接账号和数据库管理员不支持此操作")
	}
	return m.accountLiteral(name) + "@" + m.accountLiteral(host), exists, nil
}

func (m *Manager) accountLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// accountConnection 固定账号语句的转义规则，取消请求后也恢复连接的 SQL mode。
func (m *Manager) accountConnection(ctx context.Context) (*sql.Conn, func(), error) {
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	var sqlMode string
	if err = conn.QueryRowContext(ctx, "SELECT @@SESSION.sql_mode").Scan(&sqlMode); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	close := func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		if _, err := conn.ExecContext(cleanup, "SET SESSION sql_mode = ?", sqlMode); err != nil {
			_ = conn.Raw(func(interface{}) error { return driver.ErrBadConn })
		}
		_ = conn.Close()
	}
	if !strings.Contains(strings.ToUpper(sqlMode), "NO_BACKSLASH_ESCAPES") {
		mode := sqlMode
		if mode != "" {
			mode += ","
		}
		mode += "NO_BACKSLASH_ESCAPES"
		if _, err = conn.ExecContext(ctx, "SET SESSION sql_mode = ?", mode); err != nil {
			close()
			return nil, nil, err
		}
	}
	return conn, close, nil
}

// accountExec 仅执行已构造的账号语句，不负责选择业务操作。
func (m *Manager) accountExec(ctx context.Context, conn *sql.Conn, statement string) (*Result, error) {
	started := time.Now()
	result, err := conn.ExecContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, Affected: affected, ElapsedMs: time.Since(started).Milliseconds()}, nil
}
