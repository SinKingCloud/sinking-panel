package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// Permissions 读取角色在各数据库中的直接权限，不把 PUBLIC、继承及对象所有权展开为授权。
func (m *Manager) Permissions(ctx context.Context, accountName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	privileges := []string{"USAGE", "CREATE", "SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"}
	databasePrivileges := []string{"CONNECT", "CREATE", "TEMPORARY"}
	grants := []DatabasePrivileges{}
	retained := false
	result := func() interface{} {
		return map[string]interface{}{"privileges": privileges, "database_privileges": databasePrivileges, "grants": grants, "retained": retained}
	}
	if accountName == "" {
		return result(), nil
	}
	if _, exists, err := m.account(ctx, accountName); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	// 共享依赖目录只定位该角色确实有授权的数据库，不逐一连接实例中的所有数据库。
	databases, err := m.records(ctx, `SELECT d.datname AS name, d.datallowconn AND NOT d.datistemplate AS available,
EXISTS (SELECT 1 FROM pg_shdepend s WHERE s.dbid = d.oid AND s.refclassid = 'pg_authid'::regclass AND s.refobjid = r.oid AND s.deptype = 'a' AND s.classid IN ('pg_namespace'::regclass, 'pg_class'::regclass)) AS objects
FROM pg_database d CROSS JOIN pg_roles r WHERE r.rolname = $1 AND EXISTS (
SELECT 1 FROM pg_shdepend s WHERE s.refclassid = 'pg_authid'::regclass AND s.refobjid = r.oid
AND (s.dbid = d.oid OR s.dbid = 0 AND s.classid = 'pg_database'::regclass AND s.objid = d.oid)) ORDER BY d.datname`, accountName)
	if err != nil {
		return nil, err
	}
	if err = m.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_auth_members WHERE member = (SELECT oid FROM pg_roles WHERE rolname = $1)) OR EXISTS (
SELECT 1 FROM pg_shdepend WHERE refclassid = 'pg_authid'::regclass AND refobjid = (SELECT oid FROM pg_roles WHERE rolname = $1)
AND (deptype <> 'a' OR classid NOT IN ('pg_database'::regclass, 'pg_namespace'::regclass, 'pg_class'::regclass)))`, accountName).Scan(&retained); err != nil {
		return nil, err
	}
	appendGrant := func(database, schema string, objects []accountPermissionObject) {
		allowed := privileges
		if schema == "" {
			allowed = databasePrivileges
		}
		grant := DatabasePrivileges{Database: database, Schema: schema, Privileges: []string{}}
		for _, privilege := range allowed {
			count, granted := 0, 0
			for _, object := range objects {
				if object.Owned {
					retained = true
					continue
				}
				if object.Protected {
					retained = true
				}
				if schema != "" && !((privilege == "CREATE" && object.Kind == "SCHEMA") || (privilege == "USAGE" && (object.Kind == "SCHEMA" || object.Kind == "SEQUENCE")) || (privilege != "CREATE" && privilege != "USAGE" && object.Kind == "TABLE")) {
					continue
				}
				count++
				if grantable, ok := object.Privileges[privilege]; ok && !grantable {
					granted++
				}
			}
			if count > 0 && granted == count {
				grant.Privileges = append(grant.Privileges, privilege)
			} else if granted > 0 {
				retained = true
			}
		}
		for _, object := range objects {
			for privilege, grantable := range object.Privileges {
				if grantable || !slices.Contains(allowed, privilege) || object.Kind == "SEQUENCE" && privilege != "USAGE" {
					retained = true
				}
			}
		}
		if len(grant.Privileges) > 0 {
			grants = append(grants, grant)
		}
	}
	for _, record := range databases {
		database := fmt.Sprint(record["name"])
		objects, err := m.accountPermissionObjects(ctx, accountName, database, "", nil)
		if err != nil {
			return nil, err
		}
		if !m.truth(record["available"]) {
			retained = true
			continue
		}
		appendGrant(database, "", objects)
		if !m.truth(record["objects"]) {
			continue
		}
		client := m
		if database != m.database {
			client, err = NewManager(ctx, m.config, database)
			if err != nil {
				return nil, fmt.Errorf("读取数据库 %s 的角色权限失败: %w", database, err)
			}
		}
		loadErr := func() error {
			if client != m {
				defer client.Close()
			}
			schemas, err := client.records(ctx, `SELECT n.nspname AS name FROM pg_namespace n WHERE n.nspname NOT LIKE 'pg\_%' AND n.nspname <> 'information_schema' AND EXISTS (
SELECT 1 FROM pg_shdepend s WHERE s.dbid = (SELECT oid FROM pg_database WHERE datname = current_database()) AND s.refclassid = 'pg_authid'::regclass AND s.refobjid = (SELECT oid FROM pg_roles WHERE rolname = $1)
AND s.deptype = 'a' AND (s.classid = 'pg_namespace'::regclass AND s.objid = n.oid OR s.classid = 'pg_class'::regclass AND s.objid IN (SELECT oid FROM pg_class WHERE relnamespace = n.oid))) ORDER BY n.nspname`, accountName)
			if err != nil {
				return err
			}
			for _, record := range schemas {
				schema := fmt.Sprint(record["name"])
				objects, err := client.accountPermissionObjects(ctx, accountName, database, schema, nil)
				if err != nil {
					return err
				}
				appendGrant(database, schema, objects)
			}
			return nil
		}()
		if loadErr != nil {
			return nil, fmt.Errorf("读取数据库 %s 的角色权限失败: %w", database, loadErr)
		}
	}
	return result(), nil
}

// UpdatePermissions 按数据库分别提交权限差量；未提交的数据库、架构和独立对象授权保持不变。
func (m *Manager) UpdatePermissions(ctx context.Context, accountName string, grants []DatabasePrivileges) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	account, exists, err := m.account(ctx, accountName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("账号不存在，请刷新列表")
	}
	if len(grants) > 1000 {
		return nil, errors.New("一次最多修改 1000 个授权范围")
	}
	grouped := make(map[string][]DatabasePrivileges)
	seen := make(map[string]bool)
	for _, grant := range grants {
		if grant.Database == "" || len(grant.Database) > 63 || strings.ContainsRune(grant.Database, 0) || len(grant.Schema) > 63 || strings.ContainsRune(grant.Schema, 0) || grant.Database == "template0" || grant.Database == "template1" || strings.HasPrefix(grant.Schema, "pg_") || grant.Schema == "information_schema" {
			return nil, errors.New("数据库或架构名称不合法")
		}
		key := grant.Database + "\x00" + grant.Schema
		if seen[key] {
			return nil, errors.New("同一数据库和架构不能重复配置")
		}
		seen[key] = true
		allowed := []string{"USAGE", "CREATE", "SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"}
		if grant.Schema == "" {
			allowed = []string{"CONNECT", "CREATE", "TEMPORARY"}
		}
		for _, privilege := range grant.Privileges {
			if !slices.Contains(allowed, privilege) {
				return nil, errors.New("该授权范围不支持指定权限")
			}
		}
		if len(grant.Privileges) > len(allowed) {
			return nil, errors.New("数据库权限数量不合法")
		}
		grouped[grant.Database] = append(grouped[grant.Database], grant)
	}
	databases := make([]string, 0, len(grouped))
	for database := range grouped {
		databases = append(databases, database)
	}
	sort.Strings(databases)
	completed := []string{}
	retained := false
	started := time.Now()
	for _, database := range databases {
		client := m
		if database != m.database {
			client, err = NewManager(ctx, m.config, database)
			if err != nil {
				return nil, fmt.Errorf("数据库 %s 权限未保存（已完成：%s）: %w", database, strings.Join(completed, "、"), err)
			}
		}
		updateErr := func() error {
			if client != m {
				defer client.Close()
			}
			statements := []string{}
			changes := make(map[string][]accountPermissionChange)
			for _, grant := range grouped[database] {
				objects, err := client.accountPermissionObjects(ctx, accountName, database, grant.Schema, nil)
				if err != nil {
					return err
				}
				allowed := []string{"USAGE", "CREATE", "SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"}
				if grant.Schema == "" {
					allowed = []string{"CONNECT", "CREATE", "TEMPORARY"}
				}
				for _, privilege := range allowed {
					targets := []accountPermissionObject{}
					uniform := true
					for _, object := range objects {
						if object.Owned || grant.Schema != "" && !((privilege == "CREATE" && object.Kind == "SCHEMA") || (privilege == "USAGE" && (object.Kind == "SCHEMA" || object.Kind == "SEQUENCE")) || (privilege != "CREATE" && privilege != "USAGE" && object.Kind == "TABLE")) {
							continue
						}
						targets = append(targets, object)
						grantable, exists := object.Privileges[privilege]
						uniform = uniform && exists && !grantable
					}
					desired := slices.Contains(grant.Privileges, privilege)
					for _, object := range targets {
						grantable, exists := object.Privileges[privilege]
						if desired && !exists {
							statements = append(statements, "GRANT "+privilege+" ON "+object.Kind+" "+object.Name+" TO "+account)
							changes[grant.Schema] = append(changes[grant.Schema], accountPermissionChange{Name: object.Name, Kind: object.Kind, Privilege: privilege, Granted: true})
						} else if !desired && exists && (object.Protected || grantable || !uniform) {
							retained = true
						} else if !desired && uniform && exists {
							statements = append(statements, "REVOKE "+privilege+" ON "+object.Kind+" "+object.Name+" FROM "+account)
							changes[grant.Schema] = append(changes[grant.Schema], accountPermissionChange{Name: object.Name, Kind: object.Kind, Privilege: privilege})
						}
					}
				}
			}
			if len(statements) == 0 {
				return nil
			}
			tx, err := client.db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			for _, statement := range statements {
				if _, err = tx.ExecContext(ctx, statement); err != nil {
					return err
				}
			}
			// GRANT/REVOKE 在权限不足时可能仅返回 WARNING，确认实际 ACL 后才提交。
			for schema, expected := range changes {
				objects, err := client.accountPermissionObjects(ctx, accountName, database, schema, tx)
				if err != nil {
					return err
				}
				actual := make(map[string]map[string]bool, len(objects))
				for _, object := range objects {
					actual[object.Kind+"\x00"+object.Name] = object.Privileges
				}
				for _, change := range expected {
					_, granted := actual[change.Kind+"\x00"+change.Name][change.Privilege]
					if granted != change.Granted {
						return errors.New("部分权限未生效，请确认当前连接账户拥有对象授权权限")
					}
				}
			}
			return tx.Commit()
		}()
		if updateErr != nil {
			return nil, fmt.Errorf("数据库 %s 权限未保存（已完成：%s）: %w", database, strings.Join(completed, "、"), updateErr)
		}
		completed = append(completed, database)
	}
	result := map[string]interface{}{"affected": 0, "elapsed_ms": time.Since(started).Milliseconds()}
	if retained {
		result["warning"] = "权限已保存，独立对象、列级或可转授的权限已保留"
	}
	return result, nil
}

// accountPermissionObjects 读取实际 ACL；unnest(aclitem[]) 兼容旧版，不依赖较新版本的 aclexplode。
func (m *Manager) accountPermissionObjects(ctx context.Context, accountName, databaseName, schemaName string, tx *sql.Tx) ([]accountPermissionObject, error) {
	query := `SELECT d.datname AS name, 'DATABASE'::text AS kind, d.datdba = r.oid AS owned, false AS protected, CASE WHEN pg_has_role(d.datdba, 'USAGE') THEN pg_get_userbyid(d.datdba)::text ELSE CURRENT_USER::text END AS grantor, d.datacl AS acl FROM pg_database d CROSS JOIN pg_roles r WHERE r.rolname = $1 AND d.datname = $2`
	args := []interface{}{accountName, databaseName}
	if schemaName != "" {
		if _, _, err := m.accountScope(ctx, schemaName); err != nil {
			return nil, err
		}
		query = `SELECT n.nspname AS name, 'SCHEMA'::text AS kind, n.nspowner = r.oid AS owned, false AS protected, CASE WHEN pg_has_role(n.nspowner, 'USAGE') THEN pg_get_userbyid(n.nspowner)::text ELSE CURRENT_USER::text END AS grantor, n.nspacl AS acl FROM pg_namespace n CROSS JOIN pg_roles r WHERE r.rolname = $1 AND n.nspname = $2
UNION ALL SELECT c.relname AS name, CASE WHEN c.relkind = 'S' THEN 'SEQUENCE' ELSE 'TABLE' END AS kind,
c.relowner = r.oid AS owned, EXISTS (SELECT 1 FROM pg_shdepend s WHERE s.dbid = (SELECT oid FROM pg_database WHERE datname = current_database()) AND s.classid = 'pg_class'::regclass AND s.objid = c.oid AND s.objsubid > 0 AND s.refclassid = 'pg_authid'::regclass AND s.refobjid = r.oid) AS protected, CASE WHEN pg_has_role(c.relowner, 'USAGE') THEN pg_get_userbyid(c.relowner)::text ELSE CURRENT_USER::text END AS grantor, c.relacl AS acl
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace CROSS JOIN pg_roles r WHERE r.rolname = $1 AND n.nspname = $2 AND c.relkind IN ('r','v','m','f','p','S')`
		args[1] = schemaName
	}
	queryContext := m.db.QueryContext
	if tx != nil {
		queryContext = tx.QueryContext
	}
	rows, err := queryContext(ctx, "WITH objects AS ("+query+") SELECT name, kind, owned, protected, grantor, ''::text AS acl FROM objects UNION ALL SELECT name, kind, owned, protected, grantor, unnest(acl)::text FROM objects ORDER BY kind, name", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	objects := []accountPermissionObject{}
	indices := map[string]int{}
	count := 0
	letters := map[byte]string{'a': "INSERT", 'r': "SELECT", 'w': "UPDATE", 'd': "DELETE", 'D': "TRUNCATE", 'x': "REFERENCES", 't': "TRIGGER", 'U': "USAGE", 'C': "CREATE", 'T': "TEMPORARY", 'c': "CONNECT"}
	for rows.Next() {
		count++
		if count > 100000 {
			return nil, errors.New("角色权限数据超过单次读取上限")
		}
		var name, kind, grantor, acl string
		var owned, protected bool
		if err = rows.Scan(&name, &kind, &owned, &protected, &grantor, &acl); err != nil {
			return nil, err
		}
		key := kind + "\x00" + name
		index, exists := indices[key]
		if !exists {
			if len(objects) >= 10000 {
				return nil, errors.New("该架构的对象数量超过单次权限管理上限")
			}
			quoted, err := m.quote(name)
			if err != nil {
				return nil, err
			}
			if kind == "TABLE" || kind == "SEQUENCE" {
				schema, err := m.quote(schemaName)
				if err != nil {
					return nil, err
				}
				quoted = schema + "." + quoted
			}
			index = len(objects)
			indices[key] = index
			objects = append(objects, accountPermissionObject{Name: quoted, Kind: kind, Owned: owned, Protected: protected, Privileges: map[string]bool{}})
		}
		// ACL 角色名可能包含 =、/、双引号，必须按标识符引号找到真正的分隔符。
		quoted, separator := false, -1
		for i := 0; i < len(acl); i++ {
			if acl[i] == '"' {
				if quoted && i+1 < len(acl) && acl[i+1] == '"' {
					i++
				} else {
					quoted = !quoted
				}
			} else if acl[i] == '=' && !quoted {
				separator = i
				break
			}
		}
		if separator < 0 {
			continue
		}
		grantee := acl[:separator]
		if strings.HasPrefix(grantee, `"`) && strings.HasSuffix(grantee, `"`) {
			grantee = strings.ReplaceAll(grantee[1:len(grantee)-1], `""`, `"`)
		}
		if grantee != accountName {
			continue
		}
		permissions, source, _ := strings.Cut(acl[separator+1:], "/")
		if strings.HasPrefix(source, `"`) && strings.HasSuffix(source, `"`) {
			source = strings.ReplaceAll(source[1:len(source)-1], `""`, `"`)
		}
		for i := 0; i < len(permissions); i++ {
			privilege := letters[permissions[i]]
			if privilege == "" {
				privilege = string(permissions[i])
			}
			grantable := i+1 < len(permissions) && permissions[i+1] == '*'
			objects[index].Privileges[privilege] = objects[index].Privileges[privilege] || grantable || source != grantor
			if grantable {
				i++
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, errors.New("数据库或架构不存在")
	}
	return objects, nil
}
