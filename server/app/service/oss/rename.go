package oss

import (
	"context"
	utilOss "server/app/util/oss"
)

// Rename 只修改文件名称，保留原目录前缀。
func (s *service) Rename(ctx context.Context, id int64, key, name string) (*utilOss.File, error) {
	client, err := s.client(id)
	if err != nil {
		return nil, err
	}
	return client.RenameFile(ctx, key, name)
}
