package mysql

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
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
	// HEX 保留名称中的制表符等字符，不依赖 MySQL 5.7 才提供的 JSON 函数。
	output, err := m.query(ctx, `SELECT HEX(d.SCHEMA_NAME),CAST(COALESCE(SUM(t.DATA_LENGTH+t.INDEX_LENGTH),0) AS UNSIGNED) FROM information_schema.SCHEMATA d LEFT JOIN information_schema.TABLES t ON d.SCHEMA_NAME=t.TABLE_SCHEMA GROUP BY d.SCHEMA_NAME ORDER BY d.SCHEMA_NAME;`)
	if err != nil {
		return nil, err
	}
	list := []DatabaseInfo{}
	for _, row := range strings.Split(output, "\n") {
		if row == "" {
			continue
		}
		columns := strings.Split(row, "\t")
		if len(columns) != 2 {
			return nil, errors.New("数据库列表返回格式无效，请检查数据库日志")
		}
		name, nameErr := hex.DecodeString(columns[0])
		size, sizeErr := strconv.ParseInt(columns[1], 10, 64)
		if nameErr != nil || sizeErr != nil {
			return nil, errors.New("数据库列表返回格式无效，请检查数据库日志")
		}
		list = append(list, DatabaseInfo{Name: string(name), Size: size})
	}
	for i := range list {
		switch strings.ToLower(list[i].Name) {
		case "mysql", "information_schema", "performance_schema", "sys":
			list[i].Protected = true
		}
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
	case "mysql", "information_schema", "performance_schema", "sys":
		return errors.New("不能修改系统数据库")
	}
	var statement string

	statement = "CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4;"
	if action == "delete" {
		statement = "DROP DATABASE `" + name + "`;"
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
	output, err := m.query(ctx, `SELECT HEX(User),HEX(Host),Super_priv FROM mysql.user ORDER BY User,Host;`)
	if err != nil {
		return nil, err
	}
	list := []UserInfo{}
	for _, row := range strings.Split(output, "\n") {
		if row == "" {
			continue
		}
		columns := strings.Split(row, "\t")
		if len(columns) != 3 {
			return nil, errors.New("账号列表返回格式无效，请检查数据库日志")
		}
		name, nameErr := hex.DecodeString(columns[0])
		host, hostErr := hex.DecodeString(columns[1])
		if nameErr != nil || hostErr != nil {
			return nil, errors.New("账号列表返回格式无效，请检查数据库日志")
		}
		list = append(list, UserInfo{Name: string(name), Host: string(host), Protected: columns[2] == "Y"})
	}
	for i := range list {
		list[i].Protected = list[i].Protected || list[i].Name == state.Username || list[i].Name == "root" || strings.HasPrefix(list[i].Name, "mysql.")
	}
	return list, nil
}

// ChangeUser 管理业务账号；密码只通过 stdin 交给数据库客户端。
func (m *Manager) ChangeUser(ctx context.Context, action, name, password, database, host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !objectName.MatchString(name) {
		return errors.New("账号须以字母或下划线开头，仅支持字母、数字、下划线和短横线")
	}
	if action != "create" && action != "delete" && action != "password" {
		return errors.New("不支持的账号操作")
	}
	state, err := m.readState()
	if err != nil {
		return err
	}
	if name == state.Username || name == "root" {
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
		case "mysql", "information_schema", "performance_schema", "sys":
			return errors.New("业务账号不能关联系统数据库")
		}
	}
	var statement string
	// 显式固定字符串转义规则，数据库原有 SQL mode 不会改变密码的含义。
	quotedPassword := "'" + strings.ReplaceAll(password, "'", "''") + "'"

	if host == "" {
		host = "%"
	}
	if host != "%" && host != "localhost" && net.ParseIP(host) == nil {
		return errors.New("账号来源仅支持 localhost、IP 地址或 %")
	}
	account := "'" + name + "'@'" + host + "'"
	if action != "create" {
		protected, err := m.query(ctx, "SELECT COUNT(*) FROM mysql.user WHERE User='"+name+"' AND Host='"+host+"' AND Super_priv='Y';")
		if err != nil {
			return err
		}
		if strings.TrimSpace(protected) != "0" {
			return errors.New("不能修改数据库管理员账号")
		}
	}
	switch action {
	case "create":
		// MySQL 5.5/5.6 仅允许 16 字符，5.7.8 起为 32；读取系统表实际限制。
		limit, err := m.query(ctx, "SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='mysql' AND TABLE_NAME='user' AND COLUMN_NAME='User';")
		if err != nil {
			return err
		}
		maxLength, err := strconv.Atoi(limit)
		if err != nil || maxLength <= 0 {
			return errors.New("无法读取当前 MySQL 版本的账号长度限制")
		}
		if len(name) > maxLength {
			return fmt.Errorf("当前 MySQL 版本的账号名称最多 %d 位", maxLength)
		}
		exists, err := m.query(ctx, "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='"+database+"';")
		if err != nil {
			return err
		}
		if strings.TrimSpace(exists) != "1" {
			return errors.New("关联数据库不存在")
		}
		// 旧版本不存在 partial_revokes；SHOW 返回空结果时仍采用精确转义授权。
		mode, err := m.query(ctx, "SHOW GLOBAL VARIABLES LIKE 'partial_revokes';")
		if err != nil {
			return err
		}
		_, partialRevokes, _ := strings.Cut(mode, "\t")
		grantDatabase := database
		if !strings.EqualFold(partialRevokes, "ON") && partialRevokes != "1" {
			grantDatabase = strings.ReplaceAll(grantDatabase, "_", `\_`)
		}
		if _, err = m.query(ctx, "SET SESSION sql_mode='NO_BACKSLASH_ESCAPES';CREATE USER "+account+" IDENTIFIED BY "+quotedPassword+";"); err != nil {
			return err
		}
		_, err = m.query(ctx, "SET SESSION sql_mode='NO_BACKSLASH_ESCAPES';GRANT ALL PRIVILEGES ON `"+grantDatabase+"`.* TO "+account+";")
		if err != nil {
			_, cleanupErr := m.query(context.WithoutCancel(ctx), "DROP USER "+account+";")
			return errors.Join(err, cleanupErr)
		}
		return nil
	case "password":
		version, err := m.query(ctx, "SELECT VERSION();")
		if err != nil {
			return err
		}
		var major, minor, patch int
		if _, err = fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch); err != nil {
			return errors.New("无法识别当前 MySQL 版本")
		}
		statement = "ALTER USER " + account + " IDENTIFIED BY " + quotedPassword + ";"
		if major < 5 || major == 5 && (minor < 7 || minor == 7 && patch < 6) {
			statement = "SET PASSWORD FOR " + account + " = PASSWORD(" + quotedPassword + ");"
		}
	case "delete":
		statement = "DROP USER " + account + ";"
	}
	statement = "SET SESSION sql_mode='NO_BACKSLASH_ESCAPES';" + statement

	_, err = m.query(ctx, statement)
	return err
}
