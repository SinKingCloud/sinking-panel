package oss

import (
	"context"
	utilOss "server/app/util/oss"
)

// Copy 在当前存储桶内复制文件，默认保留原有元数据。
func (s *service) Copy(ctx context.Context, id int64, key, destination string, options utilOss.CopyOptions) (*utilOss.File, error) {
	client, err := s.Client(id)
	if err != nil {
		return nil, err
	}
	return client.CopyFile(ctx, key, destination, options)
}
