package redis

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	client "github.com/redis/go-redis/v9"
)

// Users 列举实例 ACL 账户，权限不能限定到单个逻辑库。
func (m *Manager) Users(ctx context.Context) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	current, err := m.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	names, err := m.client.ACLUsers(ctx).Result()
	if err != nil {
		return nil, err
	}
	if len(names) > 1000 {
		return nil, errors.New("账户数量超过 1000，请缩小管理范围")
	}
	sort.Strings(names)
	pipe := m.client.Pipeline()
	commands := make([]*client.Cmd, len(names))
	for i, name := range names {
		commands[i] = pipe.Do(ctx, "ACL", "GETUSER", name)
	}
	if _, err = pipe.Exec(ctx); err != nil && err != client.Nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, 0, len(names))
	for i, name := range names {
		value, err := commands[i].Result()
		if err == client.Nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		row, err := m.parseAccount(name, current, value)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return map[string]interface{}{"list": rows, "total": len(rows), "scope": "server", "message": "Redis ACL 权限作用于整个实例，不能限制为单个逻辑库"}, nil
}

// CreateUser 一次创建账户及其数据权限，避免创建成功但授权失败。
func (m *Manager) CreateUser(ctx context.Context, name, password string, privileges []string) (result interface{}, err error) {
	if len(password) > 1024 {
		return nil, errors.New("密码超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if password == "" || strings.ContainsRune(password, 0) {
		return nil, errors.New("密码不能为空或包含空字符")
	}
	roles, err := m.privileges(privileges)
	if err != nil {
		return nil, err
	}
	m.secrets = append(m.secrets, password)
	account, err := m.account(ctx, name)
	if err != nil {
		return nil, err
	}
	if account != nil {
		return nil, errors.New("账户已存在")
	}
	rules := []string{"on", "resetpass", ">" + password, "-@all", "+ping", "+select", "~*"}
	for _, role := range roles {
		rules = append(rules, "+"+role)
	}
	rules = append(rules, "-@dangerous", "-@admin")
	if err = m.client.ACLSetUser(ctx, name, rules...).Err(); err != nil {
		return nil, err
	}
	return m.persistAccount(ctx), nil
}

// ChangePassword 修改非保护账户的密码并持久化。
func (m *Manager) ChangePassword(ctx context.Context, name, password string) (result interface{}, err error) {
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
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	if err = m.client.ACLSetUser(ctx, name, "resetpass", ">"+password).Err(); err != nil {
		return nil, err
	}
	return m.persistAccount(ctx), nil
}

// Permissions 设置完整读写方案，以单条 ACL 命令替换权限。
func (m *Manager) Permissions(ctx context.Context, name string, privileges []string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	roles, err := m.privileges(privileges)
	if err != nil {
		return nil, err
	}
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	rules := []string{"-@all", "+ping", "+select"}
	for _, role := range roles {
		rules = append(rules, "+"+role)
	}
	rules = append(rules, "-@dangerous", "-@admin")
	if err = m.client.ACLSetUser(ctx, name, rules...).Err(); err != nil {
		return nil, err
	}
	return m.persistAccount(ctx), nil
}

// Grant 修改指定账户的读写授权。
func (m *Manager) Grant(ctx context.Context, name string, privileges []string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	roles, err := m.privileges(privileges)
	if err != nil {
		return nil, err
	}
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	rules := make([]string, 0, len(roles)+2)
	for _, role := range roles {
		rules = append(rules, "+"+role)
	}
	rules = append(rules, "-@dangerous", "-@admin")
	if err = m.client.ACLSetUser(ctx, name, rules...).Err(); err != nil {
		return nil, err
	}
	return m.persistAccount(ctx), nil
}

// Revoke 修改指定账户的读写授权。
func (m *Manager) Revoke(ctx context.Context, name string, privileges []string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	roles, err := m.privileges(privileges)
	if err != nil {
		return nil, err
	}
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	rules := make([]string, 0, len(roles)+2)
	for _, role := range roles {
		rules = append(rules, "-"+role)
	}

	if err = m.client.ACLSetUser(ctx, name, rules...).Err(); err != nil {
		return nil, err
	}
	return m.persistAccount(ctx), nil
}

// DeleteUser 删除非保护账户并持久化。
func (m *Manager) DeleteUser(ctx context.Context, name string) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if err = m.editableAccount(ctx, name); err != nil {
		return nil, err
	}
	count, err := m.client.ACLDelUser(ctx, name).Result()
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, errors.New("账户已被删除，请刷新列表")
	}
	return m.persistAccount(ctx), nil
}

func (m *Manager) currentUser(ctx context.Context) (string, error) {
	current, err := m.client.ACLWhoAmI(ctx).Result()
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "unknown command") {
		return "", errors.New("当前 Redis 未提供 ACL 账户管理，需要 Redis 6.0 或以上版本并启用 ACL 命令")
	}
	return current, err
}
func (m *Manager) account(ctx context.Context, name string) (map[string]interface{}, error) {
	if name == "" || len(name) > 128 || !utf8.ValidString(name) || strings.ContainsFunc(name, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return nil, errors.New("账户名称不能为空、含空白或控制字符，且不能超过 128 字节")
	}
	current, err := m.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	value, err := m.client.Do(ctx, "ACL", "GETUSER", name).Result()
	if err == client.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.parseAccount(name, current, value)
}
func (m *Manager) editableAccount(ctx context.Context, name string) error {
	account, err := m.account(ctx, name)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("账户不存在")
	}
	if account["protected"] == true {
		return errors.New("不能修改默认、当前连接或具有实例管理权限的受保护账户")
	}
	return nil
}
func (m *Manager) privileges(values []string) ([]string, error) {
	if len(values) == 0 || len(values) > 16 {
		return nil, errors.New("请选择有效的账户权限")
	}
	allowed := map[string]string{"read": "@read", "write": "@write"}
	roles := make([]string, 0, len(values))
	for _, value := range values {
		role, ok := allowed[value]
		if !ok {
			return nil, errors.New("Redis 账户仅支持 read、write 权限")
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (m *Manager) parseAccount(name, current string, value interface{}) (map[string]interface{}, error) {
	fields, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("Redis 返回的账户信息格式不正确")
	}
	commands := ""
	protected := name == "default" || name == current
	var inspect func([]interface{})
	inspect = func(entries []interface{}) {
		for i := 0; i+1 < len(entries); i += 2 {
			if entries[i] == "commands" {
				if text, ok := entries[i+1].(string); ok {
					if commands == "" {
						commands = text
					}
					for _, rule := range strings.Fields(text) {
						if rule == "+@all" || rule == "+@admin" || rule == "+acl" || strings.HasPrefix(rule, "+acl|") {
							protected = true
						}
					}
				}
			}
			if entries[i] == "selectors" {
				if selectors, ok := entries[i+1].([]interface{}); ok {
					for _, selector := range selectors {
						if nested, ok := selector.([]interface{}); ok {
							inspect(nested)
						}
					}
				}
			}
		}
	}
	inspect(fields)
	return map[string]interface{}{"name": name, "roles": commands, "protected": protected}, nil
}

func (m *Manager) persistAccount(ctx context.Context) interface{} {
	result := map[string]interface{}{"changed": true, "persisted": true, "scope": "server", "message": "账户操作成功"}
	// ACL 文件与 redis.conf 使用不同保存机制；持久化失败必须说明运行态已经改变。
	err := m.client.Do(ctx, "ACL", "SAVE").Err()
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "not configured") && strings.Contains(strings.ToLower(err.Error()), "acl") {
		err = m.client.ConfigRewrite(ctx).Err()
	}
	if err != nil {
		result["persisted"] = false
		result["warning"] = fmt.Sprintf("账户修改已生效，但配置持久化失败，重启后可能恢复原配置：%s", m.cleanError(err).Error())
	}
	return result
}
