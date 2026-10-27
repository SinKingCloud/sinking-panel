package mongo

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson"
	client "go.mongodb.org/mongo-driver/mongo"
)

// Users 按认证数据库和账户名列举可见账户，授权数据库与认证数据库分别返回。
func (m *Manager) Users(ctx context.Context, keyword string, page, pageSize int) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	current, err := m.currentUsers(ctx)
	if err != nil {
		return nil, err
	}
	var response struct {
		Users []account `bson:"users"`
	}
	// 全实例列举不能同时指定 showPrivileges，分页后再读取所选账户的继承角色。
	err = m.client.Database("admin").RunCommand(ctx, bson.D{{
		Key: "usersInfo",
		Value: bson.D{{
			Key:   "forAllDBs",
			Value: true,
		}},
	}, {
		Key:   "showCredentials",
		Value: false,
	}}).Decode(&response)
	scope := "instance"
	if err != nil {
		var commandError client.CommandError
		if !errors.As(err, &commandError) || commandError.Code != 13 {
			return nil, err
		}
		// 普通账户可能只能查看业务库或认证库；自身信息不要求 viewUser 权限。
		scope = "database"
		visible := make(map[[2]string]account)
		databases := map[string]bool{m.database.Name(): true}
		for _, user := range current {
			databases[user.DB] = true
		}
		for database := range databases {
			response.Users = nil
			if err = m.client.Database(database).RunCommand(ctx, bson.D{{
				Key:   "usersInfo",
				Value: 1,
			}, {
				Key:   "showCredentials",
				Value: false,
			}}).Decode(&response); err != nil {
				if !errors.As(err, &commandError) || commandError.Code != 13 {
					return nil, err
				}
				continue
			}
			for _, user := range response.Users {
				visible[[2]string{user.DB, user.User}] = user
			}
		}
		for _, user := range current {
			key := [2]string{user.DB, user.User}
			if _, exists := visible[key]; exists {
				continue
			}
			response.Users = nil
			if err = m.client.Database(user.DB).RunCommand(ctx, bson.D{{
				Key: "usersInfo",
				Value: bson.D{{
					Key:   "user",
					Value: user.User,
				}, {
					Key:   "db",
					Value: user.DB,
				}},
			}, {
				Key:   "showCredentials",
				Value: false,
			}}).Decode(&response); err != nil {
				return nil, err
			}
			for _, account := range response.Users {
				visible[[2]string{account.DB, account.User}] = account
			}
		}
		response.Users = make([]account, 0, len(visible))
		for _, user := range visible {
			response.Users = append(response.Users, user)
		}
	}
	users := make([]account, 0, len(response.Users))
	keyword = strings.ToLower(keyword)
	for _, user := range response.Users {
		if user.User != "__system" && strings.Contains(strings.ToLower(user.User), keyword) {
			users = append(users, user)
		}
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].User == users[j].User {
			return users[i].DB < users[j].DB
		}
		return users[i].User < users[j].User
	})
	if pageSize <= 0 {
		pageSize = 50
	}
	pageSize = min(pageSize, 1000)
	page = max(page, 1)
	total := len(users)
	page = min(page, max(1, (total+pageSize-1)/pageSize))
	start := min((page-1)*pageSize, total)
	users = users[start:min(start+pageSize, total)]
	if len(users) > 0 {
		selectors := make(bson.A, 0, len(users))
		for _, user := range users {
			selectors = append(selectors, bson.D{{
				Key:   "user",
				Value: user.User,
			}, {
				Key:   "db",
				Value: user.DB,
			}})
		}
		response.Users = nil
		if err = m.client.Database("admin").RunCommand(ctx, bson.D{{
			Key:   "usersInfo",
			Value: selectors,
		}, {
			Key:   "showCredentials",
			Value: false,
		}, {
			Key:   "showPrivileges",
			Value: true,
		}}).Decode(&response); err != nil {
			return nil, err
		}
		// usersInfo 的结果顺序不固定，按分页时的顺序组装。
		details := make(map[[2]string]account, len(response.Users))
		for _, user := range response.Users {
			details[[2]string{user.DB, user.User}] = user
		}
		for i, user := range users {
			if detail, ok := details[[2]string{user.DB, user.User}]; ok {
				users[i] = detail
			}
		}
	}
	rows := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		names := make([]string, 0, len(user.Roles))
		permissions := make(map[string][]string)
		for _, role := range user.Roles {
			names = append(names, role.Role+"@"+role.DB)
			switch role.Role {
			case "read", "readWrite", "dbAdmin", "userAdmin":
				permissions[role.DB] = append(permissions[role.DB], role.Role)
			}
		}
		grants := make([]DatabasePrivileges, 0, len(permissions))
		for database, privileges := range permissions {
			sort.Strings(privileges)
			grants = append(grants, DatabasePrivileges{
				Database:   database,
				Privileges: privileges,
			})
		}
		sort.Slice(grants, func(i, j int) bool { return grants[i].Database < grants[j].Database })
		sort.Strings(names)
		rows = append(rows, map[string]interface{}{
			"name":              user.User,
			"database":          user.DB,
			"roles":             strings.Join(names, ", "),
			"grants":            grants,
			"protected":         m.protectedAccount(user, current),
			"password_editable": user.DB != "$external",
		})
	}
	return map[string]interface{}{
		"list":      rows,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"scope":     scope,
		"database":  m.database.Name(),
	}, nil
}

// CreateUser 在当前认证数据库创建账户，一次授予多个数据库的常用角色。
func (m *Manager) CreateUser(ctx context.Context, name, password string, grants []DatabasePrivileges) (result interface{}, err error) {
	if len(password) > 1024 {
		return nil, errors.New("密码超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if password == "" || strings.ContainsRune(password, 0) {
		return nil, errors.New("密码不能为空或包含空字符")
	}
	m.secrets = append(m.secrets, password)
	if err = m.validateAccountName(name); err != nil {
		return nil, err
	}
	users, err := m.accounts(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(users) > 0 {
		return nil, errors.New("账户已存在")
	}
	roles, err := m.privileges(grants)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, errors.New("请选择账户的授权数据库和权限")
	}
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "createUser",
		Value: name,
	}, {
		Key:   "pwd",
		Value: password,
	}, {
		Key:   "roles",
		Value: roles,
	}}).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"changed":   true,
		"persisted": true,
		"scope":     "database",
		"database":  m.database.Name(),
		"message":   "账户操作成功",
	}, nil
}

// ChangePassword 修改指定认证库账户的密码，允许管理员和当前连接账户修改密码。
// current 以用户名和认证数据库同时匹配，便于调用方同步已保存的连接信息。
func (m *Manager) ChangePassword(ctx context.Context, name, password string) (result map[string]interface{}, current bool, err error) {
	if len(password) > 1024 {
		return nil, false, errors.New("密码超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if password == "" || strings.ContainsRune(password, 0) {
		return nil, false, errors.New("密码不能为空或包含空字符")
	}
	m.secrets = append(m.secrets, password)
	if err = m.validateAccountName(name); err != nil {
		return nil, false, err
	}
	if name == "__system" || m.database.Name() == "$external" {
		return nil, false, errors.New("系统或外部认证账户不支持设置数据库密码")
	}
	users, err := m.accounts(ctx, name)
	if err != nil {
		return nil, false, err
	}
	if len(users) == 0 {
		return nil, false, errors.New("账户不存在")
	}
	authenticated, err := m.currentUsers(ctx)
	if err != nil {
		return nil, false, err
	}
	for _, user := range authenticated {
		if user.User == name && user.DB == m.database.Name() {
			current = true
			break
		}
	}
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "updateUser",
		Value: name,
	}, {
		Key:   "pwd",
		Value: password,
	}}).Err(); err != nil {
		return nil, false, err
	}
	return map[string]interface{}{
		"changed":   true,
		"persisted": true,
		"scope":     "database",
		"database":  m.database.Name(),
		"message":   "密码修改成功",
	}, current, nil
}

// Grant 修改账户在当前数据库内的授权。
func (m *Manager) Grant(ctx context.Context, name string, privileges []string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	roles, err := m.privileges([]DatabasePrivileges{{
		Database:   m.database.Name(),
		Privileges: privileges,
	}})
	if err != nil {
		return nil, err
	}
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "grantRolesToUser",
		Value: name,
	}, {
		Key:   "roles",
		Value: roles,
	}}).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"changed":   true,
		"persisted": true,
		"scope":     "database",
		"database":  m.database.Name(),
		"message":   "账户操作成功",
	}, nil
}

// Permissions 按授权数据库返回可编辑角色，自定义和全局角色单独保留。
func (m *Manager) Permissions(ctx context.Context, name string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if err = m.validateAccountName(name); err != nil {
		return nil, err
	}
	users, err := m.accounts(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("账户不存在")
	}
	permissions := make(map[string][]string)
	other := make([]string, 0)
	for _, role := range users[0].Roles {
		switch role.Role {
		case "read", "readWrite", "dbAdmin", "userAdmin":
			permissions[role.DB] = append(permissions[role.DB], role.Role)
		default:
			other = append(other, role.Role+"@"+role.DB)
		}
	}
	grants := make([]DatabasePrivileges, 0, len(permissions))
	for database, privileges := range permissions {
		sort.Strings(privileges)
		grants = append(grants, DatabasePrivileges{
			Database:   database,
			Privileges: privileges,
		})
	}
	sort.Slice(grants, func(i, j int) bool { return grants[i].Database < grants[j].Database })
	sort.Strings(other)
	return map[string]interface{}{
		"name":        name,
		"database":    m.database.Name(),
		"grants":      grants,
		"other_roles": other,
	}, nil
}

// UpdatePermissions 替换各数据库的常用角色；自定义、全局角色不受表单修改影响。
func (m *Manager) UpdatePermissions(ctx context.Context, name string, grants []DatabasePrivileges) (result interface{}, err error) {
	if grants == nil {
		return nil, errors.New("请提供完整的数据库权限配置")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	roles, err := m.privileges(grants)
	if err != nil {
		return nil, err
	}
	if err = m.validateAccountName(name); err != nil {
		return nil, err
	}
	users, err := m.accounts(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("账户不存在")
	}
	current, err := m.currentUsers(ctx)
	if err != nil {
		return nil, err
	}
	if m.protectedAccount(users[0], current) {
		return nil, errors.New("不能修改当前连接账户或系统管理账户的权限")
	}
	for _, role := range users[0].Roles {
		switch role.Role {
		case "read", "readWrite", "dbAdmin", "userAdmin":
			continue
		}
		roles = append(roles, bson.D{{
			Key:   "role",
			Value: role.Role,
		}, {
			Key:   "db",
			Value: role.DB,
		}})
	}
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "updateUser",
		Value: name,
	}, {
		Key:   "roles",
		Value: roles,
	}}).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"changed":   true,
		"persisted": true,
		"database":  m.database.Name(),
		"message":   "权限保存成功",
	}, nil
}

// Revoke 修改账户在当前数据库内的授权。
func (m *Manager) Revoke(ctx context.Context, name string, privileges []string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	roles, err := m.privileges([]DatabasePrivileges{{
		Database:   m.database.Name(),
		Privileges: privileges,
	}})
	if err != nil {
		return nil, err
	}
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "revokeRolesFromUser",
		Value: name,
	}, {
		Key:   "roles",
		Value: roles,
	}}).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"changed":   true,
		"persisted": true,
		"scope":     "database",
		"database":  m.database.Name(),
		"message":   "账户操作成功",
	}, nil
}

// DeleteUser 删除当前数据库的非保护账户。
func (m *Manager) DeleteUser(ctx context.Context, name string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	if err = m.database.RunCommand(ctx, bson.D{{
		Key:   "dropUser",
		Value: name,
	}}).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"changed":   true,
		"persisted": true,
		"scope":     "database",
		"database":  m.database.Name(),
		"message":   "账户操作成功",
	}, nil
}

func (m *Manager) validateAccountName(name string) error {
	if name == "" || len(name) > 128 || !utf8.ValidString(name) || strings.ContainsFunc(name, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return errors.New("账户名称不能为空、含空白或控制字符，且不能超过 128 字节")
	}
	return nil
}
func (m *Manager) currentUsers(ctx context.Context) ([]account, error) {
	var status struct {
		AuthInfo struct {
			AuthenticatedUsers []account `bson:"authenticatedUsers"`
		} `bson:"authInfo"`
	}
	err := m.database.RunCommand(ctx, bson.D{{
		Key:   "connectionStatus",
		Value: 1,
	}, {
		Key:   "showPrivileges",
		Value: false,
	}}).Decode(&status)
	return status.AuthInfo.AuthenticatedUsers, err
}
func (m *Manager) protectedAccount(user account, current []account) bool {
	if user.User == "__system" {
		return true
	}
	for _, value := range current {
		if user.User == value.User && user.DB == value.DB {
			return true
		}
	}
	for _, roles := range [2][]accountRole{user.Roles, user.InheritedRoles} {
		for _, role := range roles {
			if role.DB == "admin" {
				switch role.Role {
				case "root", "__system", "userAdminAnyDatabase", "dbOwner", "userAdmin":
					return true
				}
			}
		}
	}
	return false
}
func (m *Manager) accounts(ctx context.Context, name string) ([]account, error) {
	var response struct {
		Users []account `bson:"users"`
	}
	err := m.database.RunCommand(ctx, bson.D{{
		Key: "usersInfo",
		Value: bson.D{{
			Key:   "user",
			Value: name,
		}, {
			Key:   "db",
			Value: m.database.Name(),
		}},
	}, {
		Key:   "showCredentials",
		Value: false,
	}, {
		Key:   "showPrivileges",
		Value: true,
	}}).Decode(&response)
	return response.Users, err
}
func (m *Manager) editableAccount(ctx context.Context, name string) error {
	if err := m.validateAccountName(name); err != nil {
		return err
	}
	users, err := m.accounts(ctx, name)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return errors.New("账户不存在")
	}
	if users[0].DB != m.database.Name() {
		return errors.New("不能操作其他数据库的账户")
	}
	current, err := m.currentUsers(ctx)
	if err != nil {
		return err
	}
	if m.protectedAccount(users[0], current) {
		return errors.New("不能修改当前连接账户或系统管理账户")
	}
	return nil
}

// privileges 校验授权范围并生成 MongoDB 原生角色数组；空数组表示清空可编辑角色。
func (m *Manager) privileges(grants []DatabasePrivileges) (bson.A, error) {
	if len(grants) > 1000 {
		return nil, errors.New("授权数据库数量不能超过 1000")
	}
	allowed := map[string]bool{
		"read":      true,
		"readWrite": true,
		"dbAdmin":   true,
		"userAdmin": true,
	}
	roles := bson.A{}
	seen := make(map[string]bool, len(grants))
	for _, grant := range grants {
		if grant.Database == "" || len(grant.Database) >= 64 || !utf8.ValidString(grant.Database) || strings.ContainsAny(grant.Database, "/\\. \"$*<>:|?\x00") {
			return nil, errors.New("授权数据库名称不合法")
		}
		if seen[grant.Database] {
			return nil, errors.New("授权数据库不能重复")
		}
		seen[grant.Database] = true
		if len(grant.Privileges) == 0 || len(grant.Privileges) > len(allowed) {
			return nil, errors.New("请选择有效的账户权限")
		}
		selected := make(map[string]bool, len(grant.Privileges))
		for _, role := range grant.Privileges {
			if !allowed[role] || selected[role] {
				return nil, errors.New("不支持或重复的 MongoDB 账户权限")
			}
			selected[role] = true
			roles = append(roles, bson.D{{
				Key:   "role",
				Value: role,
			}, {
				Key:   "db",
				Value: grant.Database,
			}})
		}
	}
	return roles, nil
}
