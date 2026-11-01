package oss

import (
	"context"
	utilOss "server/app/util/oss"
)

// Delete 删除当前存储桶内的文件，返回每个文件的处理结果。
func (s *service) Delete(ctx context.Context, id int64, keys []string) ([]utilOss.DeleteResult, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.DeleteFiles(ctx, keys)
}
