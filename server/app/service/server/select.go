package server

import (
	repositoryServer "server/app/repository/server"
	"server/app/util/page"
)

// Select 获取数据
func (s *service) Select(where *repositoryServer.SelectServer, queryPage *page.Query) (*page.Result[*repositoryServer.Server], error) {
	return s.repositoryServer.Select(where, queryPage)
}
