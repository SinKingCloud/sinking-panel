package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Users 分页列出角色及其在当前数据库、架构内已有对象上的有效权限。
func (m *Manager) Users(ctx context.Context, schemaName, keyword string, pageNumber, pageSize int) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	schema := schemaName
	if schema == "" {
		schema = "public"
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
	var total int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_roles WHERE rolname NOT LIKE 'pg\\_%' AND rolname ILIKE $1", "%"+keyword+"%").Scan(&total); err != nil {
		return nil, err
	}
	if total > 0 && int64((pageNumber-1)*pageSize) >= total {
		pageNumber = int((total-1)/int64(pageSize)) + 1
	}
	// 由 Postgres 判断有效权限，包含 PUBLIC、可继承角色和列级授权。
	// 先分页角色，再只检查当前架构的对象，避免逐个账户发起查询。
	privileges := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"}
	checks := make([]string, 0, len(privileges))
	for _, privilege := range privileges {
		function := "has_table_privilege"
		if privilege == "SELECT" || privilege == "INSERT" || privilege == "UPDATE" || privilege == "REFERENCES" {
			function = "has_any_column_privilege"
		}
		checks = append(checks, "EXISTS (SELECT 1 FROM objects o WHERE "+function+"(r.oid, o.oid, '"+privilege+"')) AS "+strings.ToLower(privilege)+"_priv")
	}
	records, err := m.records(ctx, `WITH accounts AS (
    SELECT oid, rolname, rolsuper, rolcreatedb, rolcreaterole, rolcanlogin
    FROM pg_roles WHERE rolname NOT LIKE 'pg\_%' AND rolname ILIKE $2 ORDER BY rolname LIMIT $3 OFFSET $4
), objects AS (
    SELECT c.oid, c.relowner FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = $1 AND c.relkind IN ('r', 'v', 'm', 'f', 'p')
)
SELECT r.rolname AS name, r.rolsuper, r.rolcreatedb, r.rolcreaterole, r.rolcanlogin,
    r.rolname = CURRENT_USER OR r.rolname = SESSION_USER AS current_account,
    NOT r.rolsuper OR r.rolname = CURRENT_USER OR r.rolname = SESSION_USER
        OR (SELECT rolsuper FROM pg_roles WHERE rolname = CURRENT_USER) AS password_editable,
    has_database_privilege(r.oid, current_database(), 'CONNECT') AS connect_priv,
    has_database_privilege(r.oid, current_database(), 'CREATE') AS create_db_priv,
    has_database_privilege(r.oid, current_database(), 'TEMPORARY') AS temporary_priv,
    has_schema_privilege(r.oid, (SELECT oid FROM pg_namespace WHERE nspname = $1), 'USAGE') AS schema_usage,
    has_schema_privilege(r.oid, (SELECT oid FROM pg_namespace WHERE nspname = $1), 'CREATE') AS schema_create,
    EXISTS (SELECT 1 FROM objects o WHERE pg_has_role(r.oid, o.relowner, 'USAGE')) AS object_owner,
    `+strings.Join(checks, ",\n    ")+`
FROM accounts r ORDER BY r.rolname`, schema, "%"+keyword+"%", pageSize, (pageNumber-1)*pageSize)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		roles := make([]string, 0)
		for _, attribute := range []struct{ key, label string }{
			{"rolsuper", "SUPERUSER"}, {"rolcreatedb", "CREATEDB"}, {"rolcreaterole", "CREATEROLE"}, {"rolcanlogin", "LOGIN"},
			{"connect_priv", "CONNECT"}, {"temporary_priv", "TEMPORARY"}, {"schema_usage", "USAGE"}, {"object_owner", "OWNER"},
		} {
			if m.truth(record[attribute.key]) {
				roles = append(roles, attribute.label)
			}
		}
		if m.truth(record["create_db_priv"]) || m.truth(record["schema_create"]) {
			roles = append(roles, "CREATE")
		}
		for _, privilege := range privileges {
			if m.truth(record[strings.ToLower(privilege)+"_priv"]) {
				roles = append(roles, privilege)
			}
		}
		name := fmt.Sprint(record["name"])
		protected := m.truth(record["rolsuper"]) || m.truth(record["current_account"]) || name == "postgres" || strings.HasPrefix(name, "pg_")
		list = append(list, map[string]interface{}{"name": record["name"], "roles": strings.Join(roles, ", "), "protected": protected, "password_editable": m.truth(record["password_editable"])})
	}
	return map[string]interface{}{"list": list, "total": total, "page": pageNumber, "page_size": pageSize}, nil
}

// CreateUser 创建登录角色，不自动授予对象权限。
func (m *Manager) CreateUser(ctx context.Context, accountName, accountPassword string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	account, exists, err := m.account(ctx, accountName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("该账号已存在")
	}
	if accountPassword == "" || len(accountPassword) > 1024 || strings.ContainsRune(accountPassword, 0) {
		return nil, errors.New("密码应为 1 至 1024 字节，且不能包含空字符")
	}

	// E 字符串明确反斜线语义，不依赖 standard_conforming_strings 设置。
	password := "E'" + strings.ReplaceAll(strings.ReplaceAll(accountPassword, `\`, `\\`), "'", "''") + "'"
	started := time.Now()
	result, err := m.db.ExecContext(ctx, "CREATE ROLE "+account+" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE PASSWORD "+password)
	if err != nil {
		return nil, errors.New("账号或密码设置失败，请检查账号权限、名称和密码策略")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, Affected: affected, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// ChangePassword 修改角色密码，由 Postgres 校验权限；返回是否修改了当前登录账户。
func (m *Manager) ChangePassword(ctx context.Context, accountName, accountPassword string) (out *Result, current bool, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, false, err
	}
	if accountName == "" || len(accountName) > 63 || strings.ContainsRune(accountName, 0) {
		return nil, false, errors.New("账号名称应为 1 至 63 字节，且不能包含空字符")
	}
	if accountName == "public" || strings.HasPrefix(accountName, "pg_") {
		return nil, false, errors.New("不能修改系统角色")
	}
	account, err := m.quote(accountName)
	if err != nil {
		return nil, false, err
	}
	if err = m.db.QueryRowContext(ctx, "SELECT rolname = SESSION_USER FROM pg_roles WHERE rolname = $1", accountName).Scan(&current); errors.Is(err, sql.ErrNoRows) {
		return nil, false, errors.New("账号不存在，请刷新列表")
	} else if err != nil {
		return nil, false, err
	}
	if accountPassword == "" || len(accountPassword) > 1024 || strings.ContainsRune(accountPassword, 0) {
		return nil, false, errors.New("密码应为 1 至 1024 字节，且不能包含空字符")
	}

	// E 字符串明确反斜线语义，不依赖 standard_conforming_strings 设置。
	password := "E'" + strings.ReplaceAll(strings.ReplaceAll(accountPassword, `\`, `\\`), "'", "''") + "'"
	started := time.Now()
	result, err := m.db.ExecContext(ctx, "ALTER ROLE "+account+" PASSWORD "+password)
	if err != nil {
		return nil, false, errors.New("账号或密码设置失败，请检查账号权限、名称和密码策略")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, Affected: affected, ElapsedMs: time.Since(started).Milliseconds()}, current, nil
}

// Grant 在事务中授予当前数据库、模式和现有表的权限。
func (m *Manager) Grant(ctx context.Context, accountName, schemaName string, permissions []string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	account, exists, err := m.account(ctx, accountName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	targets, err := m.accountPrivileges(ctx, schemaName, permissions)
	if err != nil {
		return nil, err
	}
	objects := make([]string, 0, len(targets))
	for object := range targets {
		objects = append(objects, object)
	}
	sort.Strings(objects)
	statements := make([]string, 0, len(objects))
	for _, object := range objects {
		statements = append(statements, "GRANT "+strings.Join(targets[object], ", ")+" ON "+object+" TO "+account)
	}
	return m.accountTransaction(ctx, statements)
}

// Revoke 在事务中撤回当前数据库、模式和现有表的权限。
func (m *Manager) Revoke(ctx context.Context, accountName, schemaName string, permissions []string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	account, exists, err := m.account(ctx, accountName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	targets, err := m.accountPrivileges(ctx, schemaName, permissions)
	if err != nil {
		return nil, err
	}
	objects := make([]string, 0, len(targets))
	for object := range targets {
		objects = append(objects, object)
	}
	sort.Strings(objects)
	statements := make([]string, 0, len(objects))
	for _, object := range objects {
		statements = append(statements, "REVOKE "+strings.Join(targets[object], ", ")+" ON "+object+" FROM "+account)
	}
	return m.accountTransaction(ctx, statements)
}

// DeleteUser 撤回当前范围授权后删除角色，不删除角色拥有的数据。
func (m *Manager) DeleteUser(ctx context.Context, accountName, schemaName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	account, exists, err := m.account(ctx, accountName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	database, schema, err := m.accountScope(ctx, schemaName)
	if err != nil {
		return nil, err
	}
	statements := make([]string, 0, 5)
	// 不使用 DROP OWNED；角色仍拥有对象时，删除失败并回滚本次撤权。
	objects := []string{"DATABASE " + database, "SCHEMA " + schema, "ALL TABLES IN SCHEMA " + schema, "ALL SEQUENCES IN SCHEMA " + schema}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	if version < 90000 {
		tables, sequences, err := m.accountRelations(ctx, schemaName)
		if err != nil {
			return nil, err
		}
		objects = append(objects[:2], tables...)
		objects = append(objects, sequences...)
	}
	for _, object := range objects {
		statements = append(statements, "REVOKE ALL PRIVILEGES ON "+object+" FROM "+account)
	}
	statements = append(statements, "DROP ROLE "+account)
	return m.accountTransaction(ctx, statements)
}

// account 校验角色名称并读取保护状态，创建与修改由公开方法分别决定。
func (m *Manager) account(ctx context.Context, accountName string) (string, bool, error) {
	name := accountName
	if name == "" || len(name) > 63 || strings.ContainsRune(name, 0) {
		return "", false, errors.New("账号名称应为 1 至 63 字节，且不能包含空字符")
	}
	if name == "postgres" || name == "public" || strings.HasPrefix(name, "pg_") {
		return "", false, errors.New("不能修改系统角色")
	}
	var super, current bool
	err := m.db.QueryRowContext(ctx, "SELECT rolsuper, rolname = CURRENT_USER OR rolname = SESSION_USER FROM pg_roles WHERE rolname = $1", name).Scan(&super, &current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	exists := err == nil
	if exists && (super || current) {
		return "", false, errors.New("当前连接账号和数据库超级用户仅支持查看")
	}
	account, err := m.quote(name)
	return account, exists, err
}

func (m *Manager) accountScope(ctx context.Context, schemaName string) (string, string, error) {
	schema := schemaName
	if schema == "" {
		schema = "public"
	}
	if strings.HasPrefix(schema, "pg_") || schema == "information_schema" || m.database == "template0" || m.database == "template1" {
		return "", "", errors.New("不能在系统模式或模板数据库中管理业务权限")
	}
	quotedSchema, err := m.quote(schema)
	if err != nil {
		return "", "", err
	}
	database, err := m.quote(m.database)
	if err != nil {
		return "", "", err
	}
	var exists bool
	if err = m.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)", schema).Scan(&exists); err != nil {
		return "", "", err
	}
	if !exists {
		return "", "", errors.New("指定模式不存在")
	}
	return database, quotedSchema, nil
}

// accountPrivileges 按 Postgres 对象范围分配经过允许列表校验的权限。
func (m *Manager) accountPrivileges(ctx context.Context, schemaName string, permissions []string) (map[string][]string, error) {
	database, schema, err := m.accountScope(ctx, schemaName)
	if err != nil {
		return nil, err
	}
	allowed := map[string]bool{"CONNECT": true, "USAGE": true, "SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true, "TRUNCATE": true, "REFERENCES": true, "TRIGGER": true, "CREATE": true}
	if len(permissions) == 0 || len(permissions) > len(allowed) {
		return nil, errors.New("请选择有效的数据库权限")
	}
	targets := make(map[string][]string)
	seen := make(map[string]bool)
	for _, privilege := range permissions {
		privilege = strings.ToUpper(strings.TrimSpace(privilege))
		if !allowed[privilege] {
			return nil, errors.New("不支持的数据库权限")
		}
		if seen[privilege] {
			continue
		}
		seen[privilege] = true
		switch privilege {
		case "CONNECT":
			targets["DATABASE "+database] = append(targets["DATABASE "+database], privilege)
		case "CREATE":
			targets["DATABASE "+database] = append(targets["DATABASE "+database], privilege)
			targets["SCHEMA "+schema] = append(targets["SCHEMA "+schema], privilege)
		case "USAGE":
			targets["SCHEMA "+schema] = append(targets["SCHEMA "+schema], privilege)
			targets["ALL SEQUENCES IN SCHEMA "+schema] = append(targets["ALL SEQUENCES IN SCHEMA "+schema], privilege)
		default:
			targets["ALL TABLES IN SCHEMA "+schema] = append(targets["ALL TABLES IN SCHEMA "+schema], privilege)
		}
	}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	if version < 90000 {
		// PostgreSQL 8.4 没有 ALL TABLES/SEQUENCES IN SCHEMA，批量读取后逐对象授权。
		tables, sequences, err := m.accountRelations(ctx, schemaName)
		if err != nil {
			return nil, err
		}
		for target, relations := range map[string][]string{"ALL TABLES IN SCHEMA " + schema: tables, "ALL SEQUENCES IN SCHEMA " + schema: sequences} {
			privileges := targets[target]
			delete(targets, target)
			if len(privileges) == 0 {
				continue
			}
			for _, relation := range relations {
				targets[relation] = append(targets[relation], privileges...)
			}
		}
	}
	return targets, nil
}

// accountRelations 一次读取旧版本模式内的表和序列，不按对象发起元数据请求。
func (m *Manager) accountRelations(ctx context.Context, schemaName string) ([]string, []string, error) {
	if schemaName == "" {
		schemaName = "public"
	}
	records, err := m.records(ctx, "SELECT c.relname AS name, c.relkind AS kind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relkind IN ('r','v','S') ORDER BY c.relname", schemaName)
	if err != nil {
		return nil, nil, err
	}
	schema, err := m.quote(schemaName)
	if err != nil {
		return nil, nil, err
	}
	tables, sequences := []string{}, []string{}
	for _, record := range records {
		name, err := m.quote(fmt.Sprint(record["name"]))
		if err != nil {
			return nil, nil, err
		}
		if record["kind"] == "S" {
			sequences = append(sequences, "SEQUENCE "+schema+"."+name)
		} else {
			tables = append(tables, "TABLE "+schema+"."+name)
		}
	}
	return tables, sequences, nil
}

// accountTransaction 原子执行已构造的授权语句，不接收业务操作标识。
func (m *Manager) accountTransaction(ctx context.Context, statements []string) (*Result, error) {
	started := time.Now()
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

// SetPrivileges 原子更新当前数据库和架构的权限方案，不修改角色身份和其他数据库。
func (m *Manager) SetPrivileges(ctx context.Context, accountName, schemaName string, permissions []string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	account, exists, err := m.account(ctx, accountName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	desired := make(map[string][]string)
	if len(permissions) > 0 {
		desired, err = m.accountPrivileges(ctx, schemaName, permissions)
		if err != nil {
			return nil, err
		}
	}
	all, err := m.accountPrivileges(ctx, schemaName, []string{"CONNECT", "USAGE", "CREATE", "SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"})
	if err != nil {
		return nil, err
	}
	objects := make([]string, 0, len(all))
	for object := range all {
		objects = append(objects, object)
	}
	sort.Strings(objects)
	statements := make([]string, 0, len(objects)*2)
	for _, object := range objects {
		statements = append(statements, "REVOKE "+strings.Join(all[object], ", ")+" ON "+object+" FROM "+account)
		if len(desired[object]) > 0 {
			statements = append(statements, "GRANT "+strings.Join(desired[object], ", ")+" ON "+object+" TO "+account)
		}
	}
	return m.accountTransaction(ctx, statements)
}
