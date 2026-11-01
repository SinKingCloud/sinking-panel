package oss

import (
	"context"
	utilOss "server/app/util/oss"
)

// Move 将文件移动到当前存储桶内的完整目标路径，复制成功后删除原文件。
func (s *service) Move(ctx context.Context, id int64, key, destination string) (*utilOss.File, error) {
	client, err := s.client(id)
	if err != nil {
		return nil, err
	}
	return client.MoveFile(ctx, key, destination)
}
