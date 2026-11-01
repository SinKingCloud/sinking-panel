package oss

import (
	"encoding/json"
	"errors"
	"server/app/model"
	"strings"
	"unicode/utf8"
)

// Create 保存基本信息，并按关联密钥的厂商初始化连接配置字段。
func (s *service) Create(data *model.Oss) error {
	if data == nil {
		return errors.New("存储信息不能为空")
	}
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" || utf8.RuneCountInString(data.Name) > 100 {
		return errors.New("存储名称应为 1 至 100 个字符")
	}
	if data.SecretId <= 0 {
		return errors.New("请选择云密钥")
	}
	secret, err := s.repositorySecret.FindById(data.SecretId)
	if err != nil {
		return err
	}
	if secret == nil {
		return errors.New("云密钥不存在")
	}
	_, config, err := s.formatConfig("{}", secret.Provider)
	if err != nil {
		return err
	}
	content, err := json.Marshal(config)
	if err != nil {
		return err
	}
	data.Config = string(content)
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.repository.Create(data)
}
