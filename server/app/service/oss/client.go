package oss

import (
	"context"
	"errors"
	utilOss "server/app/util/oss"

	"gorm.io/gorm"
)

// newClient 读取密钥并解析对应厂商配置，返回客户端和可保存的连接配置。
func (s *service) newClient(secretId int64, content string, tx ...*gorm.DB) (utilOss.Interface, map[string]string, error) {
	if secretId <= 0 {
		return nil, nil, errors.New("请选择云密钥")
	}
	secret, err := s.repositorySecret.FindById(secretId, tx...)
	if err != nil {
		return nil, nil, err
	}
	if secret == nil {
		return nil, nil, errors.New("云密钥不存在")
	}
	config, fields, err := s.formatConfig(content, secret.Provider)
	if err != nil {
		return nil, nil, err
	}
	if config.Config().Bucket == "" {
		return nil, nil, errors.New("请填写存储桶")
	}
	client, err := utilOss.NewFromSecret(secret.Provider, secret.Data, config)
	if err != nil {
		return nil, nil, err
	}
	return client, fields, nil
}

// Client 读取已保存的连接并创建统一客户端，调用者直接使用文件及分片上传 API。
func (s *service) Client(id int64) (utilOss.Interface, error) {
	if id <= 0 {
		return nil, errors.New("存储 ID 不合法")
	}
	data, err := s.repository.FindById(id)
	if err != nil {
		return nil, err
	}
	if data == nil || data.Id <= 0 {
		return nil, errors.New("对象存储不存在")
	}
	client, _, err := s.newClient(data.SecretId, data.Config)
	return client, err
}

// Test 使用传入的密钥和连接配置检查存储桶，无需先保存存储记录。
func (s *service) Test(ctx context.Context, secretId int64, config string) error {
	client, _, err := s.newClient(secretId, config)
	if err != nil {
		return err
	}
	return client.HeadBucket(ctx)
}
