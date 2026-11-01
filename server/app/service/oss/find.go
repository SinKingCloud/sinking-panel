package oss

import "errors"

// FindById 查询对象存储详情，返回解析后的连接配置。
func (s *service) FindById(id int64) (*Oss, error) {
	if id <= 0 {
		return nil, errors.New("存储 ID 不合法")
	}
	data, err := s.repository.FindById(id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("对象存储不存在")
	}
	secret, err := s.repositorySecret.FindById(data.SecretId)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, errors.New("云密钥不存在")
	}
	_, config, err := s.formatConfig(data.Config, secret.Provider)
	if err != nil {
		return nil, err
	}
	return &Oss{
		Id:         data.Id,
		Name:       data.Name,
		SecretId:   data.SecretId,
		Config:     config,
		CreateTime: data.CreateTime,
		UpdateTime: data.UpdateTime,
	}, nil
}
