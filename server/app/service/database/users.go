package database

import (
	"context"
	"encoding/json"
	"errors"
	"server/app/enum/database_kind"
	repositoryDatabase "server/app/repository/database"
	"server/app/service/database/mongo"
	"server/app/service/database/mysql"
	"server/app/service/database/postgres"
	"server/app/service/database/sqlserver"
	"strconv"
	"strings"
)

// ChangeMongoPassword 修改账户密码；当前认证账户同时更新面板保存的凭据。
func (s *service) ChangeMongoPassword(ctx context.Context, id int64, databaseName, name, password string) (interface{}, error) {
	connection, err := s.FindConnection(id, database_kind.MongoDB)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	var config mongo.Config
	if err = json.Unmarshal([]byte(connection.Config), &config); err != nil {
		return nil, errors.New("连接配置不合法")
	}
	m, err := mongo.NewManager(ctx, connection.Config, databaseName)
	if err != nil {
		return nil, err
	}
	defer m.Close()
	data, current, err := m.ChangePassword(ctx, name, password)
	if err != nil {
		return nil, err
	}
	data["current"] = current
	data["connection_updated"] = false
	if !current {
		return data, nil
	}
	// 远端密码已经生效，本地配置更新失败时明确提示，避免重复修改远端密码。
	data["warning"] = "密码已修改，但实例连接信息未保存，请在配置中填写新密码"
	authSource := config.AuthSource
	if authSource == "" {
		authSource = "admin"
	}
	latest, err := s.FindConnection(id, database_kind.MongoDB)
	if err != nil || latest.Config != connection.Config || config.User != name || data["database"] != authSource {
		return data, nil
	}
	config.Password = password
	content, err := json.Marshal(config)
	if err != nil {
		return data, nil
	}
	// 明文密码直接随连接配置加密，避免将密码内容误识别为接口返回的密文。
	encrypted := s.encrypt(strconv.FormatInt(id, 10), string(content))
	if encrypted == "" {
		return data, nil
	}
	if err = s.repository.UpdateById(id, &repositoryDatabase.UpdateDatabase{Config: &encrypted}); err != nil {
		return data, nil
	}
	delete(data, "warning")
	data["connection_updated"] = true
	return data, nil
}

// ChangeMysqlPassword 修改账户密码；当前连接账户同时更新面板保存的凭据。
func (s *service) ChangeMysqlPassword(ctx context.Context, id int64, databaseName, name, host, password string) (interface{}, error) {
	connection, err := s.FindConnection(id, database_kind.MySQL)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	var config mysql.Config
	if err = json.Unmarshal([]byte(connection.Config), &config); err != nil {
		return nil, errors.New("连接配置不合法")
	}
	m, err := mysql.NewManager(ctx, connection.Config, databaseName)
	if err != nil {
		return nil, err
	}
	defer m.Close()
	result, current, err := m.ChangePassword(ctx, name, host, password)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{"affected": result.Affected, "elapsed_ms": result.ElapsedMs, "current": current, "connection_updated": false}
	if !current {
		return data, nil
	}
	// 已完成远端改密，本地保存失败只返回明确提示，避免误导用户重复改密。
	data["warning"] = "密码已修改，但实例连接信息未保存，请在配置中填写新密码"
	latest, err := s.FindConnection(id, database_kind.MySQL)
	if err != nil || latest.Config != connection.Config || config.User != name {
		return data, nil
	}
	config.Password = password
	content, err := json.Marshal(config)
	if err != nil {
		return data, nil
	}
	// 内部已取得明文配置，直接加密保存，避免把以 enc:v1: 开头的新密码误当作提交密文。
	encrypted := s.encrypt(strconv.FormatInt(id, 10), string(content))
	if encrypted == "" {
		return data, nil
	}
	if err = s.repository.UpdateById(id, &repositoryDatabase.UpdateDatabase{Config: &encrypted}); err != nil {
		return data, nil
	}
	delete(data, "warning")
	data["connection_updated"] = true
	return data, nil
}

// ChangePostgresPassword 修改角色密码；当前登录账户同时更新面板保存的凭据。
func (s *service) ChangePostgresPassword(ctx context.Context, id int64, databaseName, name, password string) (interface{}, error) {
	connection, err := s.FindConnection(id, database_kind.Postgres)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	var config postgres.Config
	if err = json.Unmarshal([]byte(connection.Config), &config); err != nil {
		return nil, errors.New("连接配置不合法")
	}
	m, err := postgres.NewManager(ctx, connection.Config, databaseName)
	if err != nil {
		return nil, err
	}
	defer m.Close()
	result, current, err := m.ChangePassword(ctx, name, password)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{"affected": result.Affected, "elapsed_ms": result.ElapsedMs, "current": current, "connection_updated": false}
	if !current {
		return data, nil
	}
	// 远端密码已生效，配置变动或保存失败时提示补填，不返回改密失败。
	data["warning"] = "密码已修改，但实例连接信息未保存，请在配置中填写新密码"
	latest, err := s.FindConnection(id, database_kind.Postgres)
	if err != nil || latest.Config != connection.Config || config.User != "" && config.User != name {
		return data, nil
	}
	config.User = name
	config.Password = password
	content, err := json.Marshal(config)
	if err != nil {
		return data, nil
	}
	encrypted := s.encrypt(strconv.FormatInt(id, 10), string(content))
	if encrypted == "" {
		return data, nil
	}
	if err = s.repository.UpdateById(id, &repositoryDatabase.UpdateDatabase{Config: &encrypted}); err != nil {
		return data, nil
	}
	delete(data, "warning")
	data["connection_updated"] = true
	return data, nil
}

// ChangeSqlServerPassword 修改登录密码；当前连接账户同时更新面板保存的凭据。
func (s *service) ChangeSqlServerPassword(ctx context.Context, id int64, databaseName, name, password string) (interface{}, error) {
	connection, err := s.FindConnection(id, database_kind.SqlServer)
	if err != nil {
		return nil, err
	}
	if len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return nil, errors.New("数据库名称不合法")
	}
	var config sqlserver.Config
	if err = json.Unmarshal([]byte(connection.Config), &config); err != nil {
		return nil, errors.New("连接配置不合法")
	}
	m, err := sqlserver.NewManager(ctx, connection.Config, databaseName)
	if err != nil {
		return nil, err
	}
	defer m.Close()
	result, current, err := m.ChangePassword(ctx, name, password)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{"affected": result.Affected, "elapsed_ms": result.ElapsedMs, "current": current, "connection_updated": false}
	if !current {
		return data, nil
	}
	// 远端密码已生效；保留原登录名（可能大小写不同），仅在配置未变时更新密码。
	data["warning"] = "密码已修改，但实例连接信息未保存，请在配置中填写新密码"
	latest, err := s.FindConnection(id, database_kind.SqlServer)
	if err != nil || latest.Config != connection.Config {
		return data, nil
	}
	config.Password = password
	content, err := json.Marshal(config)
	if err != nil {
		return data, nil
	}
	encrypted := s.encrypt(strconv.FormatInt(id, 10), string(content))
	if encrypted == "" {
		return data, nil
	}
	if err = s.repository.UpdateById(id, &repositoryDatabase.UpdateDatabase{Config: &encrypted}); err != nil {
		return data, nil
	}
	delete(data, "warning")
	data["connection_updated"] = true
	return data, nil
}
