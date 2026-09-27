package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"regexp"
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
	statement := `var r=db.getSiblingDB('admin').runCommand({listDatabases:1}); if(!r.ok) throw Error(r.errmsg); print(JSON.stringify(r.databases.map(function(x){return {name:x.name,size:Number(x.sizeOnDisk)};})));`

	output, err := m.query(ctx, statement)
	if err != nil {
		return nil, err
	}
	list := []DatabaseInfo{}
	if err = json.Unmarshal([]byte(strings.TrimSpace(output)), &list); err != nil {
		return nil, errors.New("数据库列表返回格式无效，请检查数据库日志")
	}
	for i := range list {
		switch strings.ToLower(list[i].Name) {
		case "admin", "local", "config":
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
	case "admin", "local", "config":
		return errors.New("不能修改系统数据库")
	}
	var statement string

	encoded, _ := json.Marshal(name)
	// MongoDB 只有首次创建集合后才持久化数据库，空集合不占用业务文档。
	statement = fmt.Sprintf(`var n=%s;if(db.getMongo().getDBNames().indexOf(n)!==-1)throw Error('数据库已存在');var r=db.getSiblingDB(n).createCollection('_sinking_init');if(!r.ok)throw Error(r.errmsg);`, encoded)
	if action == "delete" {
		statement = fmt.Sprintf(`var r=db.getSiblingDB(%s).dropDatabase();if(!r.ok)throw Error(r.errmsg);`, encoded)
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
	// 历史版本没有 usersInfo.forAllDBs；只投影账号和角色，认证散列不进入返回结果。
	statement := `var users=db.getSiblingDB('admin').system.users.find({}, {user:1,db:1,roles:1,_id:0}).toArray();print(JSON.stringify(users.map(function(x){return {name:x.user,database:x.db,roles:x.roles.map(function(y){return y.role+'@'+y.db;}).join(', '),protected:x.roles.some(function(y){return ['root','userAdminAnyDatabase','dbAdminAnyDatabase','userAdmin','dbOwner'].indexOf(y.role)!==-1;})};})));`

	output, err := m.query(ctx, statement)
	if err != nil {
		return nil, err
	}
	list := []UserInfo{}
	if err = json.Unmarshal([]byte(strings.TrimSpace(output)), &list); err != nil {
		return nil, errors.New("账号列表返回格式无效，请检查数据库日志")
	}
	for i := range list {
		list[i].Protected = list[i].Protected || list[i].Name == state.Username || list[i].Name == "root"
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
	if name == state.Username || name == "root" {
		return errors.New("安装管理员和系统账号不能在业务账号中修改")
	}
	if action != "delete" && (len(password) < 8 || len(password) > 256 || strings.ContainsAny(password, "\x00\r\n")) {
		return errors.New("密码需要 8 至 256 位，不能包含换行或空字符")
	}
	if !objectName.MatchString(database) {
		return errors.New("请选择有效的业务数据库")
	}
	if database != "" {
		switch strings.ToLower(database) {
		case "admin", "local", "config":
			return errors.New("业务账号不能关联系统数据库")
		}
	}
	var statement string

	n, _ := json.Marshal(name)
	p, _ := json.Marshal(password)
	d, _ := json.Marshal(database)
	if action != "create" {
		protected, err := m.query(ctx, fmt.Sprintf(`var u=db.getSiblingDB(%s).getUser(%s);print(u&&u.roles.some(function(x){return ['root','userAdminAnyDatabase','dbAdminAnyDatabase','userAdmin','dbOwner'].indexOf(x.role)!==-1;})?'1':'0');`, d, n))
		if err != nil {
			return err
		}
		if strings.TrimSpace(protected) != "0" {
			return errors.New("不能修改数据库管理员账号")
		}
	}
	switch action {
	case "create":
		statement = fmt.Sprintf(`db.getSiblingDB(%s).createUser({user:%s,pwd:%s,roles:[{role:'readWrite',db:%s}]});`, d, n, p, d)
	case "password":
		statement = fmt.Sprintf(`db.getSiblingDB(%s).updateUser(%s,{pwd:%s});`, d, n, p)
	case "delete":
		statement = fmt.Sprintf(`if(!db.getSiblingDB(%s).dropUser(%s))throw Error('账号不存在');`, d, n)
	}

	_, err = m.query(ctx, statement)
	return err
}
