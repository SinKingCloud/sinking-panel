package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"
)

// accountQuery 根据 SID 关联数据库用户和登录，保留系统与当前账号的保护信息。
const accountQuery = `SELECT p.principal_id AS id, p.name, l.name AS login_name,
CASE WHEN p.principal_id <= 4 OR l.principal_id <= 1 OR l.principal_id IS NULL
OR p.type <> 'S' OR l.type <> 'S' OR p.name = USER_NAME()
OR l.sid = SUSER_SID() OR l.sid = SUSER_SID(ORIGINAL_LOGIN()) OR l.name LIKE '##%'
OR IS_SRVROLEMEMBER('sysadmin', l.name) = 1 OR IS_SRVROLEMEMBER('securityadmin', l.name) = 1
OR IS_ROLEMEMBER('db_owner', p.name) = 1 OR IS_ROLEMEMBER('db_securityadmin', p.name) = 1
THEN 1 ELSE 0 END AS protected
FROM sys.database_principals p LEFT JOIN sys.server_principals l ON l.sid = p.sid
WHERE p.type IN ('S', 'U', 'G')`

// Users 列出当前库的账户及权限。
func (m *Manager) Users(ctx context.Context, keyword string, pageNumber, pageSize int) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
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
	// 内部数据库用户不展示；sa 和当前 SQL 登录可能只映射到 dbo，单独补入供修改密码。
	accounts := `SELECT a.id, a.name COLLATE DATABASE_DEFAULT AS name, a.login_name COLLATE DATABASE_DEFAULT AS login_name, a.protected
FROM (` + accountQuery + ` AND p.principal_id > 4 AND p.name NOT LIKE '##%' AND (l.name IS NULL OR l.name NOT LIKE '##%')) AS a
UNION ALL SELECT -l.principal_id, l.name COLLATE DATABASE_DEFAULT, l.name COLLATE DATABASE_DEFAULT, 1
FROM sys.server_principals l WHERE l.type = 'S' AND l.name NOT LIKE '##%'
AND (l.sid = 0x01 OR l.sid = SUSER_SID(ORIGINAL_LOGIN()))
AND NOT EXISTS (SELECT 1 FROM sys.database_principals p WHERE p.sid = l.sid AND p.principal_id > 4 AND p.type IN ('S', 'U', 'G') AND p.name NOT LIKE '##%')`
	// 列表预判可改密范围，实际修改仍由 SQL Server 验证权限与密码策略。
	filtered := `SELECT a.*, CASE WHEN l.type = 'S' AND l.name NOT LIKE '##%' AND (
l.sid = SUSER_SID(ORIGINAL_LOGIN()) OR HAS_PERMS_BY_NAME(NULL, NULL, 'CONTROL SERVER') = 1
OR (l.sid <> 0x01 AND IS_SRVROLEMEMBER('sysadmin', l.name) = 0
AND (HAS_PERMS_BY_NAME(NULL, NULL, 'ALTER ANY LOGIN') = 1 OR HAS_PERMS_BY_NAME(l.name, 'LOGIN', 'ALTER') = 1)
AND NOT EXISTS (SELECT 1 FROM sys.server_permissions permission WHERE permission.grantee_principal_id = l.principal_id
AND permission.class = 100 AND permission.permission_name = 'CONTROL SERVER' AND permission.state IN ('G', 'W')))
) THEN 1 ELSE 0 END AS password_editable
FROM (` + accounts + `) AS a LEFT JOIN sys.server_principals l ON a.login_name IS NOT NULL AND l.sid = SUSER_SID(a.login_name)
WHERE a.name LIKE @p1 OR a.login_name LIKE @p1`
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ("+filtered+") AS accounts", "%"+keyword+"%").Scan(&total); err != nil {
		return nil, err
	}
	if total > 0 && int64((pageNumber-1)*pageSize) >= total {
		pageNumber = int((total-1)/int64(pageSize)) + 1
	}
	records, err := m.records(ctx, "SELECT * FROM (SELECT accounts.*,ROW_NUMBER() OVER(ORDER BY name,id) AS row_number FROM ("+filtered+") AS accounts) AS paged WHERE row_number>@p2 AND row_number<=@p3 ORDER BY row_number", "%"+keyword+"%", (pageNumber-1)*pageSize, pageNumber*pageSize)

	if err != nil {
		return nil, err
	}
	// 角色名与权限描述的系统视图列可能使用不同排序规则，合并前统一为当前数据库。
	roles, err := m.records(ctx, `SELECT d.member_principal_id AS id, r.name COLLATE DATABASE_DEFAULT AS role
FROM sys.database_role_members d JOIN sys.database_principals r ON r.principal_id = d.role_principal_id
UNION ALL SELECT grantee_principal_id, state_desc COLLATE DATABASE_DEFAULT + N' ' + permission_name COLLATE DATABASE_DEFAULT FROM sys.database_permissions WHERE class = 0
ORDER BY id, role`)
	if err != nil {
		return nil, err
	}
	permissions := make(map[string][]string)
	for _, role := range roles {
		id := fmt.Sprint(role["id"])
		permissions[id] = append(permissions[id], fmt.Sprint(role["role"]))
	}
	list := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		list = append(list, map[string]interface{}{
			"name": record["name"], "login_name": record["login_name"],
			"roles": strings.Join(permissions[fmt.Sprint(record["id"])], ", "), "protected": m.truth(record["protected"]),
			"password_editable": m.truth(record["password_editable"]),
		})
	}
	return map[string]interface{}{"list": list, "total": total, "page": pageNumber, "page_size": pageSize}, nil
}

// CreateUser 在事务内创建登录和当前数据库用户。
func (m *Manager) CreateUser(ctx context.Context, accountName, accountPassword string, permissions []string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := m.userName(accountName); err != nil {
		return nil, err
	}
	if err := m.password(accountPassword); err != nil {
		return nil, err
	}
	privileges, err := m.userPrivileges(permissions)
	if err != nil {
		return nil, err
	}
	name, err := m.quote(accountName)
	if err != nil {
		return nil, err
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 密码参数先绑定，再由 SqlServer 生成转义后的 Unicode 字面量。
	statement := "DECLARE @statement nvarchar(max); SET @statement = N'CREATE LOGIN ' + QUOTENAME(@p1) + N' WITH PASSWORD = N' + QUOTENAME(@p2, '''') + N', DEFAULT_DATABASE = ' + QUOTENAME(DB_NAME()); EXEC sys.sp_executesql @statement;"
	if _, err = tx.ExecContext(ctx, statement, accountName, accountPassword); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "CREATE USER "+name+" FOR LOGIN "+name); err != nil {
		return nil, err
	}
	if len(privileges) > 0 {
		if _, err = tx.ExecContext(ctx, "GRANT "+strings.Join(privileges, ", ")+" TO "+name); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]interface{}{"name": accountName, "action": "create"}, nil
}

// ChangePassword 修改 SQL 登录密码，返回是否为当前连接登录，供调用方同步连接凭据。
func (m *Manager) ChangePassword(ctx context.Context, accountName, accountPassword string) (out *Result, current bool, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, false, err
	}
	if accountName == "" || len(utf16.Encode([]rune(accountName))) > 128 || strings.ContainsAny(accountName, "\x00\r\n") {
		return nil, false, errors.New("登录名称须为 1 至 128 位，不能包含换行或空字符")
	}
	if strings.HasPrefix(accountName, "##") {
		return nil, false, errors.New("不能修改内部系统登录")
	}
	// secrets 的第一项是建立连接时使用的密码；普通登录修改自身密码时需要 OLD_PASSWORD。
	oldPassword := ""
	if len(m.secrets) > 0 {
		oldPassword = m.secrets[0]
	}
	if err = m.password(accountPassword); err != nil {
		return nil, false, err
	}

	started := time.Now()
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	var login string
	err = tx.QueryRowContext(ctx, `SELECT name, CASE WHEN sid = SUSER_SID(ORIGINAL_LOGIN()) THEN 1 ELSE 0 END
FROM sys.server_principals WHERE name = @p1 AND type = 'S' AND name NOT LIKE '##%'`, accountName).Scan(&login, &current)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, errors.New("SQL 登录不存在或当前账户无权访问")
	} else if err != nil {
		return nil, false, err
	}
	statement := "DECLARE @statement nvarchar(max); SET @statement = N'ALTER LOGIN ' + QUOTENAME(@p1) + N' WITH PASSWORD = N' + QUOTENAME(@p2, '''');"
	arguments := []interface{}{login, accountPassword}
	if current {
		m.secrets = append(m.secrets, strings.ReplaceAll(oldPassword, "'", "''"))
		statement += " SET @statement = @statement + N' OLD_PASSWORD = N' + QUOTENAME(@p3, '''');"
		arguments = append(arguments, oldPassword)
	}
	statement += " EXEC sys.sp_executesql @statement;"
	if _, err = tx.ExecContext(ctx, statement, arguments...); err != nil {
		return nil, false, err
	}

	if err = tx.Commit(); err != nil {
		return nil, false, err
	}

	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, current, nil
}

// Grant 在当前数据库内授予指定权限。
func (m *Manager) Grant(ctx context.Context, accountName string, permissions []string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	privileges, err := m.userPrivileges(permissions)
	if err != nil {
		return nil, err
	}
	if len(privileges) == 0 {
		return nil, errors.New("请选择数据库权限")
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	name, _, err := m.user(ctx, tx, accountName)
	if err != nil {
		return nil, err
	}
	quoted, err := m.quote(name)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "GRANT "+strings.Join(privileges, ", ")+" TO "+quoted); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]interface{}{"name": name, "action": "grant"}, nil
}

// Revoke 在当前数据库内撤销指定权限。
func (m *Manager) Revoke(ctx context.Context, accountName string, permissions []string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	privileges, err := m.userPrivileges(permissions)
	if err != nil {
		return nil, err
	}
	if len(privileges) == 0 {
		return nil, errors.New("请选择数据库权限")
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	name, _, err := m.user(ctx, tx, accountName)
	if err != nil {
		return nil, err
	}
	quoted, err := m.quote(name)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "REVOKE "+strings.Join(privileges, ", ")+" FROM "+quoted); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]interface{}{"name": name, "action": "revoke"}, nil
}

// DeleteUser 在同一事务中清理登录的全部数据库用户，再删除登录。
func (m *Manager) DeleteUser(ctx context.Context, accountName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userName, loginName, err := m.user(ctx, tx, accountName)
	if err != nil {
		return nil, err
	}
	login, err := m.quote(loginName)
	if err != nil {
		return nil, err
	}

	if err = m.deleteUserMappings(ctx, tx, loginName); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "DROP LOGIN "+login); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": userName, "action": "delete"}, nil
}

// deleteUserMappings 按 SID 清理实际用户名称，保留数据对象及其他账号的显式权限。
func (m *Manager) deleteUserMappings(ctx context.Context, tx *sql.Tx, loginName string) error {
	// 没有完整元数据可见性时，不能确认全部关联，也不能安全转移所有权。
	var visible int
	if err := tx.QueryRowContext(ctx, "SELECT CASE WHEN HAS_PERMS_BY_NAME(NULL, NULL, 'VIEW ANY DEFINITION') = 1 AND HAS_PERMS_BY_NAME(NULL, NULL, 'VIEW ANY DATABASE') = 1 THEN 1 ELSE 0 END").Scan(&visible); err != nil {
		return err
	}
	if visible != 1 {
		return errors.New("删除登录需要查看全部数据库的账号关联，请使用有 VIEW ANY DEFINITION 和 VIEW ANY DATABASE 权限的账号")
	}
	var owner sql.NullString
	if err := tx.QueryRowContext(ctx, "SELECT SUSER_SNAME(0x01)").Scan(&owner); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT name, state, HAS_DBACCESS(name) FROM sys.databases ORDER BY name")
	if err != nil {
		return err
	}
	databases := make([]string, 0)
	for rows.Next() {
		var database string
		var state int
		var accessible sql.NullInt64
		if err = rows.Scan(&database, &state, &accessible); err != nil {
			_ = rows.Close()
			return err
		}
		if state != 0 || !accessible.Valid || accessible.Int64 != 1 {
			_ = rows.Close()
			return fmt.Errorf("无法确认数据库 %s 的账号关联，未删除登录", database)
		}
		databases = append(databases, database)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	// sp_executesql 的数据库限定名只改变此批次的上下文，不污染连接的默认数据库。
	statement := `SET XACT_ABORT ON;
DECLARE @sid varbinary(85), @id int, @name sysname, @kind nvarchar(20), @target nvarchar(776), @sql nvarchar(max);
SET @sid=SUSER_SID(@login);
IF @sid IS NULL BEGIN RAISERROR(N'登录不存在，未删除账号',16,1); RETURN; END;
IF COALESCE(HAS_PERMS_BY_NAME(DB_NAME(),'DATABASE','VIEW DEFINITION'),0)<>1 BEGIN RAISERROR(N'无法查看完整数据库用户和所有权，未删除账号',16,1); RETURN; END;
IF EXISTS(SELECT 1 FROM sys.database_principals WHERE sid=@sid AND ((principal_id<=4 AND NOT(principal_id=1 AND EXISTS(SELECT 1 FROM sys.databases WHERE database_id=DB_ID() AND owner_sid=@sid))) OR type<>'S' OR name LIKE N'##%'))
BEGIN RAISERROR(N'不能删除系统数据库用户',16,1); RETURN; END;
IF EXISTS(SELECT 1 FROM sys.databases WHERE database_id=DB_ID() AND owner_sid=@sid)
BEGIN
 IF @sa=N'' BEGIN RAISERROR(N'无法读取系统登录，不能转移数据库所有权',16,1); RETURN; END;
 SET @sql=N'ALTER AUTHORIZATION ON DATABASE::'+QUOTENAME(DB_NAME())+N' TO '+QUOTENAME(@sa); EXEC sys.sp_executesql @sql;
END;
DECLARE users_to_remove CURSOR LOCAL FAST_FORWARD FOR SELECT principal_id,name FROM sys.database_principals WHERE sid=@sid AND principal_id>4 ORDER BY principal_id;
OPEN users_to_remove; FETCH NEXT FROM users_to_remove INTO @id,@name;
WHILE @@FETCH_STATUS=0
BEGIN
 IF EXISTS(SELECT 1 FROM sys.database_permissions p WHERE p.grantee_principal_id NOT IN(SELECT principal_id FROM sys.database_principals WHERE sid=@sid) AND
  (p.grantor_principal_id=@id OR
   (p.class=3 AND EXISTS(SELECT 1 FROM sys.schemas s WHERE s.schema_id=p.major_id AND s.principal_id=@id)) OR
   (p.class=4 AND EXISTS(SELECT 1 FROM sys.database_principals r WHERE r.principal_id=p.major_id AND r.owning_principal_id=@id)) OR
   (p.class=1 AND EXISTS(SELECT 1 FROM sys.objects o JOIN sys.schemas s ON s.schema_id=o.schema_id WHERE o.object_id=p.major_id AND (o.principal_id=@id OR (o.principal_id IS NULL AND s.principal_id=@id)))) OR
   (p.class=6 AND EXISTS(SELECT 1 FROM sys.types t JOIN sys.schemas s ON s.schema_id=t.schema_id WHERE t.user_type_id=p.major_id AND (t.principal_id=@id OR (t.principal_id IS NULL AND s.principal_id=@id)))) OR
   (p.class=10 AND EXISTS(SELECT 1 FROM sys.xml_schema_collections x JOIN sys.schemas s ON s.schema_id=x.schema_id WHERE x.xml_collection_id=p.major_id AND (x.principal_id=@id OR (x.principal_id IS NULL AND s.principal_id=@id))))))
 BEGIN RAISERROR(N'转移账号所有权会影响其他账号的显式权限，请先调整授权后删除',16,1); RETURN; END;
 DECLARE ownership_to_transfer CURSOR LOCAL FAST_FORWARD FOR
 SELECT N'OBJECT',QUOTENAME(s.name) COLLATE DATABASE_DEFAULT+N'.'+QUOTENAME(o.name) COLLATE DATABASE_DEFAULT FROM sys.objects o JOIN sys.schemas s ON s.schema_id=o.schema_id WHERE o.principal_id=@id
 UNION ALL SELECT N'ROLE',QUOTENAME(name) COLLATE DATABASE_DEFAULT FROM sys.database_principals WHERE type='R' AND owning_principal_id=@id
 UNION ALL SELECT N'SCHEMA',QUOTENAME(name) COLLATE DATABASE_DEFAULT FROM sys.schemas WHERE principal_id=@id;
 OPEN ownership_to_transfer; FETCH NEXT FROM ownership_to_transfer INTO @kind,@target;
 WHILE @@FETCH_STATUS=0
 BEGIN
  SET @sql=N'ALTER AUTHORIZATION ON '+@kind+N'::'+@target+N' TO [dbo]'; EXEC sys.sp_executesql @sql;
  FETCH NEXT FROM ownership_to_transfer INTO @kind,@target;
 END;
 CLOSE ownership_to_transfer; DEALLOCATE ownership_to_transfer;
 SET @sql=N'DROP USER '+QUOTENAME(@name); EXEC sys.sp_executesql @sql;
 FETCH NEXT FROM users_to_remove INTO @id,@name;
END;
CLOSE users_to_remove; DEALLOCATE users_to_remove;`
	for _, database := range databases {
		quoted, quoteErr := m.quote(database)
		if quoteErr != nil {
			return quoteErr
		}
		if _, err := tx.ExecContext(ctx, "EXEC "+quoted+".sys.sp_executesql @p1, N'@login sysname, @sa sysname', @p2, @p3", statement, loginName, owner.String); err != nil {
			return fmt.Errorf("清理数据库 %s 的账号关联失败，未删除登录: %w", database, err)
		}
	}
	return nil
}

// userName 校验 SQL 登录账户名，系统保留名称不允许变更。
func (m *Manager) userName(name string) error {
	if name == "" || len(utf16.Encode([]rune(name))) > 128 || strings.ContainsAny(name, "\x00\r\n\\") {
		return errors.New("账号名称须为 1 至 128 位，不能包含换行、空字符或反斜杠")
	}
	switch strings.ToLower(name) {
	case "sa", "dbo", "guest", "sys", "information_schema", "public":
		return errors.New("不能修改系统账号")
	}
	if strings.HasPrefix(name, "##") {
		return errors.New("不能修改系统账号")
	}
	return nil
}

// user 读取实际数据库用户和登录名，拒绝受保护账号。
func (m *Manager) user(ctx context.Context, tx *sql.Tx, name string) (string, string, error) {
	if err := m.userName(name); err != nil {
		return "", "", err
	}
	var id int64
	var login sql.NullString
	var protected int
	err := tx.QueryRowContext(ctx, accountQuery+" AND p.name = @p1", name).Scan(&id, &name, &login, &protected)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", errors.New("账号不存在或当前账户无权访问")
	}
	if err != nil {
		return "", "", err
	}
	if protected != 0 || !login.Valid {
		return "", "", errors.New("系统账号、管理员和当前连接账号仅支持查看")
	}
	return name, login.String, nil
}

// password 校验密码并加入错误脱敏列表。
func (m *Manager) password(value string) error {
	length := len(utf16.Encode([]rune(value)))
	if length < 8 || length > 128 || strings.ContainsAny(value, "\x00\r\n") {
		return errors.New("密码须为 8 至 128 位，不能包含换行或空字符")
	}
	m.secrets = append(m.secrets, value, strings.ReplaceAll(value, "'", "''"))
	return nil
}

// userPrivileges 仅接受当前库可以授予的业务权限。
func (m *Manager) userPrivileges(values []string) ([]string, error) {
	allowed := map[string]bool{"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true, "EXECUTE": true, "CREATE TABLE": true, "CREATE VIEW": true}
	privileges := make([]string, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if !allowed[value] {
			return nil, errors.New("不支持的数据库权限")
		}
		if !seen[value] {
			privileges = append(privileges, value)
			seen[value] = true
		}
	}
	return privileges, nil
}

// SetPrivileges 在同一事务内替换可视化表单管理的数据库权限。
func (m *Manager) SetPrivileges(ctx context.Context, accountName string, permissions []string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	privileges, err := m.userPrivileges(permissions)
	if err != nil {
		return nil, err
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	name, _, err := m.user(ctx, tx, accountName)
	if err != nil {
		return nil, err
	}
	quoted, err := m.quote(name)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "REVOKE SELECT, INSERT, UPDATE, DELETE, EXECUTE, CREATE TABLE, CREATE VIEW FROM "+quoted); err != nil {
		return nil, err
	}
	if len(privileges) > 0 {
		if _, err = tx.ExecContext(ctx, "GRANT "+strings.Join(privileges, ", ")+" TO "+quoted); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": name, "action": "permissions"}, nil
}
