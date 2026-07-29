package server

import (
	"errors"
	"server/app/enum/server_auth_type"
	"server/app/model"
	"server/app/util/str"
)

// Create 插入数据
func (s *service) Create(data *model.Server) error {
	if data == nil {
		return errors.New("服务器数据不能为空")
	}
	if _, ok := server_auth_type.Map()[data.AuthType]; !ok {
		return errors.New("服务器验证类型不合法")
	}
	data.Id = str.GetSnowWorkIns().GetId()
	return s.repositoryServer.Create(data)
}
