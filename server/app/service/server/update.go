package server

import (
	"errors"
	"server/app/enum/server_auth_type"
	repositoryServer "server/app/repository/server"
)

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryServer.UpdateServer) error {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	if data.AuthType != nil {
		value, ok := data.AuthType.(int)
		if !ok {
			return errors.New("服务器验证类型不合法")
		}
		if _, ok = server_auth_type.Map()[value]; !ok {
			return errors.New("服务器验证类型不合法")
		}
	}
	return s.repositoryServer.UpdateByIds(ids, data)
}
