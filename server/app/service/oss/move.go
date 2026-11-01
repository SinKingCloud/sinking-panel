package oss

import (
	"context"
	utilOss "server/app/util/oss"
)

// Move 移动或重命名当前存储桶内的文件，复制成功后删除原文件。
func (s *service) Move(ctx context.Context, id int64, key, destination string) (*utilOss.File, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.MoveFile(ctx, key, destination)
}
