package oss

import (
	repositoryOss "server/app/repository/oss"
	"server/app/util/page"
)

// Select 分页查询对象存储基本信息。
func (s *service) Select(where *repositoryOss.SelectOss, query *page.Query) (*page.Result[*repositoryOss.Oss], error) {
	return s.repository.Select(where, query)
}
