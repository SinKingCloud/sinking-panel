package server

import (
	"errors"
	"server/app/enum/server_auth_type"
	repositoryServer "server/app/repository/server"
	"server/app/util/page"
)

// Select 获取数据
func (s *service) Select(where *repositoryServer.SelectServer, queryPage *page.Query) (*page.Result[*repositoryServer.Server], error) {
	if where != nil && where.AuthType != nil {
		if _, ok := server_auth_type.Map()[*where.AuthType]; !ok {
			return nil, errors.New("服务器验证类型不合法")
		}
	}
	return s.repositoryServer.Select(where, queryPage)
}
