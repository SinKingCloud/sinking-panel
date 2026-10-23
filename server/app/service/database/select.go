package database

import (
	"encoding/json"
	"errors"
	"server/app/enum/database_type"
	"server/app/model"
	repositoryDatabase "server/app/repository/database"
	"server/app/util/page"
	"strconv"
)

// FindById 返回实例详情及完整配置，密码等凭据以密文返回。
func (s *service) FindById(id int64) (*Database, error) {
	data, err := s.repository.FindById(id)
	if err != nil || data == nil {
		return nil, err
	}
	stored := data.Config
	if data.Config != "" {
		data.Config = s.decrypt(strconv.FormatInt(data.Id, 10), data.Config)
		if data.Config == "" {
			return nil, errors.New("连接配置解密失败")
		}
	}
	config, err := s.Config(data.Type, data.Config, data.Id, stored, "encrypt")
	if err != nil {
		return nil, err
	}
	return &Database{
		Id:         data.Id,
		Name:       data.Name,
		Type:       data.Type,
		TypeId:     data.TypeId,
		Config:     config,
		CreateTime: data.CreateTime,
		UpdateTime: data.UpdateTime,
	}, nil
}

// Select 分页查询实例。
func (s *service) Select(where *repositoryDatabase.SelectDatabase, query *page.Query) (*page.Result[*repositoryDatabase.Database], error) {
	if where != nil && where.Type != nil {
		if _, ok := database_type.Map()[*where.Type]; !ok {
			return nil, errors.New("数据库类型不合法")
		}
	}
	return s.repository.Select(where, query)
}

// FindConnection 按实例 ID 校验类型并解密连接配置，不读取或登记远端数据库。
func (s *service) FindConnection(id int64, databaseType int) (*model.Database, error) {
	if id <= 0 {
		return nil, errors.New("实例 ID 不合法")
	}
	if databaseType != 0 {
		if _, ok := database_type.Map()[databaseType]; !ok {
			return nil, errors.New("数据库类型不合法")
		}
	}
	data, err := s.repository.FindById(id)
	if err != nil {
		return nil, err
	}
	if data == nil || data.Id <= 0 {
		return nil, errors.New("实例不存在")
	}
	if _, ok := database_type.Map()[data.Type]; !ok {
		return nil, errors.New("数据库类型不合法")
	}
	if databaseType != 0 && data.Type != databaseType {
		return nil, errors.New("数据库实例类型与当前操作不匹配")
	}
	if data.Config == "" {
		return nil, errors.New("实例尚未配置连接信息，请先编辑实例")
	}
	connection := *data
	connection.Config = s.decrypt(strconv.FormatInt(data.Id, 10), data.Config)
	if connection.Config == "" {
		return nil, errors.New("连接配置解密失败")
	}
	config, err := s.Config(connection.Type, connection.Config, 0, "", "")
	if err != nil {
		return nil, err
	}
	content, _ := json.Marshal(config)
	connection.Config = string(content)
	return &connection, nil
}
