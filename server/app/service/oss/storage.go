package oss

import (
	"context"
	utilOss "server/app/util/oss"
	"time"
)

// Check 校验已保存的存储连接和关联密钥，不发起网络请求。
func (s *service) Check(id int64) error {
	_, err := s.Client(id)
	return err
}

// StatFile 读取当前存储桶中文件的元数据。
func (s *service) StatFile(ctx context.Context, id int64, key string) (*utilOss.File, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.StatFile(ctx, key)
}

// CreateDirectory 创建当前存储桶中的目录前缀。
func (s *service) CreateDirectory(ctx context.Context, id int64, key string) (*utilOss.File, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.CreateDirectory(ctx, key)
}

// SetMetadata 更新文件元数据和内容响应头。
func (s *service) SetMetadata(ctx context.Context, id int64, key string, options utilOss.PutOptions) (*utilOss.File, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.SetMetadata(ctx, key, options)
}

// Sign 为当前存储桶的文件生成限时访问地址。
func (s *service) Sign(ctx context.Context, id int64, key, method string, expiry time.Duration, options ...utilOss.SignOptions) (*utilOss.SignedURL, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.Sign(ctx, key, method, expiry, options...)
}
