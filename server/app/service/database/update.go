package database

import (
	"encoding/json"
	"errors"
	"server/app/enum/type_module"
	repositoryDatabase "server/app/repository/database"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Update 修改实例信息，数据库类型在登记后保持不变。
func (s *service) Update(id int64, data *repositoryDatabase.UpdateDatabase) error {
	if id <= 0 || data == nil {
		return errors.New("实例更新参数不合法")
	}
	if data.Name != nil {
		name := strings.TrimSpace(*data.Name)
		if name == "" || utf8.RuneCountInString(name) > 100 {
			return errors.New("实例名称应为 1 至 100 个字符")
		}
		data.Name = &name
	}
	if data.TypeId != nil {
		if *data.TypeId < 0 {
			return errors.New("分类不合法")
		}
		if *data.TypeId > 0 {
			category, err := s.typeService.FindById(*data.TypeId)
			if err != nil {
				return err
			}
			if category == nil || category.Module != type_module.Database {
				return errors.New("数据库分类不存在")
			}
		}
	}
	server, err := s.repository.FindById(id)
	if err != nil {
		return err
	}
	if server == nil {
		return errors.New("实例不存在")
	}
	if data.Config != nil {
		if strings.TrimSpace(*data.Config) == "" {
			return errors.New("请填写连接配置")
		}
		config, err := s.Config(server.Type, *data.Config, id, server.Config, "decrypt")
		if err != nil {
			return err
		}
		content, _ := json.Marshal(config)
		encrypted := s.encrypt(strconv.FormatInt(id, 10), string(content))
		if encrypted == "" {
			return errors.New("连接配置加密失败")
		}
		data.Config = &encrypted
	}
	return s.repository.UpdateById(id, data)
}
