package pgsql

import (
	"context"
	"encoding/csv"
	"errors"

	"regexp"
	"strconv"
	"strings"
)

// 管理页创建的名称限制为常见标识符；已有特殊名称仍可列出，避免将输入拼成命令。
var objectName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,62}$`)

// DatabaseInfo 是数据库列表的轻量信息，大小单位为字节。
type DatabaseInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Owner     string `json:"owner,omitempty"`
	Protected bool   `json:"protected"`
}

// UserInfo 不返回密码，Host 用于 MySQL，Database 用于 MongoDB 和受管 PostgreSQL 账号。
type UserInfo struct {
	Name      string `json:"name"`
	Host      string `json:"host,omitempty"`
	Database  string `json:"database,omitempty"`
	Roles     string `json:"roles,omitempty"`
	Protected bool   `json:"protected"`
}

// Databases 读取实际数据库列表，不在面板数据库中重复保存。
func (m *Manager) Databases(ctx context.Context) ([]DatabaseInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	statement := `COPY (SELECT datname,pg_database_size(oid),pg_get_userbyid(datdba) FROM pg_database WHERE NOT datistemplate ORDER BY datname) TO STDOUT WITH CSV;`

	output, err := m.query(ctx, statement)
	if err != nil {
		return nil, err
	}
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		return nil, errors.New("数据库列表返回格式无效，请检查数据库日志")
	}
	list := make([]DatabaseInfo, 0, len(rows))
	for _, row := range rows {
		if len(row) != 3 {
			return nil, errors.New("数据库列表返回格式无效，请检查数据库日志")
		}
		size, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			return nil, errors.New("数据库大小返回格式无效，请检查数据库日志")
		}
		item := DatabaseInfo{Name: row[0], Size: size, Owner: row[2]}
		switch strings.ToLower(item.Name) {
		case "postgres", "template0", "template1":
			item.Protected = true
		}
		list = append(list, item)
	}
	return list, nil
}

// ChangeDatabase 创建或删除数据库，系统数据库不可操作。
func (m *Manager) ChangeDatabase(ctx context.Context, action, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !objectName.MatchString(name) {
		return errors.New("数据库名称须以字母或下划线开头，仅支持字母、数字、下划线和短横线，最多 63 位")
	}
	if action != "create" && action != "delete" {
		return errors.New("不支持的数据库操作")
	}
	switch strings.ToLower(name) {
	case "postgres", "template0", "template1":
		return errors.New("不能修改系统数据库")
	}
	var statement string

	statement = `CREATE DATABASE "` + name + `";`
	if action == "delete" {
		statement = `DROP DATABASE "` + name + `";`
	}

	_, err := m.query(ctx, statement)
	return err
}

// Users 列出账号和权限归属，系统账号和安装管理员仅供查看。
func (m *Manager) Users(ctx context.Context) ([]UserInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state, err := m.readState()
	if err != nil {
		return nil, err
	}
	protected := "rolsuper OR rolcreaterole"
	if m.versionNumber(state.Version) >= 90100 {
		protected += " OR rolreplication"
	}
	statement := `COPY (SELECT rolname,(` + protected + `),CASE WHEN shobj_description(oid,'pg_authid') LIKE 'sinking-panel:%' THEN substr(shobj_description(oid,'pg_authid'),15) ELSE '' END FROM pg_roles WHERE rolcanlogin ORDER BY rolname) TO STDOUT WITH CSV;`

	output, err := m.query(ctx, statement)
	if err != nil {
		return nil, err
	}
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		return nil, errors.New("账号列表返回格式无效，请检查数据库日志")
	}
	list := make([]UserInfo, 0, len(rows))
	for _, row := range rows {
		if len(row) != 3 {
			return nil, errors.New("账号列表返回格式无效，请检查数据库日志")
		}
		list = append(list, UserInfo{Name: row[0], Database: row[2], Protected: row[1] == "t" || row[0] == state.Username || row[0] == "root" || strings.HasPrefix(row[0], "pg_")})
	}
	return list, nil
}

// ChangeUser 管理业务账号；密码只通过 stdin 交给数据库客户端。
func (m *Manager) ChangeUser(ctx context.Context, action, name, password, database, host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !objectName.MatchString(name) {
		return errors.New("账号须以字母或下划线开头，仅支持字母、数字、下划线和短横线，最多 63 位")
	}
	if action != "create" && action != "delete" && action != "password" {
		return errors.New("不支持的账号操作")
	}
	state, err := m.readState()
	if err != nil {
		return err
	}
	if name == state.Username || name == "root" || strings.HasPrefix(name, "pg_") {
		return errors.New("安装管理员和系统账号不能在业务账号中修改")
	}
	if action != "delete" && (len(password) < 8 || len(password) > 256 || strings.ContainsAny(password, "\x00\r\n")) {
		return errors.New("密码需要 8 至 256 位，不能包含换行或空字符")
	}
	if (action == "create" || database != "") && !objectName.MatchString(database) {
		return errors.New("请选择有效的业务数据库")
	}
	if database != "" {
		switch strings.ToLower(database) {
		case "postgres", "template0", "template1":
			return errors.New("业务账号不能关联系统数据库")
		}
	}
	var statement string
	// 固定字符串转义规则，避免数据库原有配置改变密码的含义。
	quotedPassword := "'" + strings.ReplaceAll(password, "'", "''") + "'"

	account := `"` + name + `"`
	if action != "create" {
		condition := "rolsuper OR rolcreaterole"
		if m.versionNumber(state.Version) >= 90100 {
			condition += " OR rolreplication"
		}
		protected, err := m.query(ctx, "SELECT count(*) FROM pg_roles WHERE rolname='"+name+"' AND ("+condition+");")
		if err != nil {
			return err
		}
		if strings.TrimSpace(protected) != "0" {
			return errors.New("不能修改数据库管理员账号")
		}
	}
	switch action {
	case "create":
		// 记录授权目标，删除账号时可以撤销这些权限；不触碰该账号拥有的业务对象。
		statement = "\\connect " + database + "\nBEGIN;SET standard_conforming_strings=on;CREATE ROLE " + account + " LOGIN PASSWORD " + quotedPassword + ";COMMENT ON ROLE " + account + " IS 'sinking-panel:" + database + "';GRANT CONNECT ON DATABASE \"" + database + "\" TO " + account + ";GRANT USAGE,CREATE ON SCHEMA public TO " + account + ";COMMIT;"
	case "password":
		statement = "SET standard_conforming_strings=on;ALTER ROLE " + account + " PASSWORD " + quotedPassword + ";"
	case "delete":
		statement = "DROP ROLE " + account + ";"
		if database != "" {
			exists, err := m.query(ctx, "SELECT count(*) FROM pg_database WHERE datname='"+database+"';")
			if err != nil {
				return err
			}
			if strings.TrimSpace(exists) == "1" {
				statement = "\\connect " + database + "\nBEGIN;REVOKE ALL PRIVILEGES ON DATABASE \"" + database + "\" FROM " + account + ";REVOKE ALL PRIVILEGES ON SCHEMA public FROM " + account + ";" + statement + "COMMIT;"
			}
		}
	}

	_, err = m.query(ctx, statement)
	return err
}
