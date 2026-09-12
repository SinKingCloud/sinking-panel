package server

import (
	"errors"
	"strconv"

	"server/app/constant"
	"server/app/enum/server_auth_type"
	repositoryServer "server/app/repository/server"
)

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryServer.UpdateServer) error {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	if data.AuthType != nil {
		if _, ok := server_auth_type.Map()[*data.AuthType]; !ok {
			return errors.New("服务器验证类型不合法")
		}
	}
	if len(ids) == 1 && ids[0] == 0 {
		configs := make(map[string]string)
		if data.Ip != nil {
			configs[constant.SshIP] = *data.Ip
		}
		if data.Port != nil {
			configs[constant.SshPort] = strconv.Itoa(*data.Port)
		}
		if data.User != nil {
			configs[constant.SshUser] = *data.User
		}
		if data.AuthType != nil {
			configs[constant.SshAuthType] = strconv.Itoa(*data.AuthType)
		}
		if data.Password != nil {
			configs[constant.SshPassword] = *data.Password
		}
		if data.Name != nil {
			configs[constant.SshName] = *data.Name
		}
		if len(configs) == 0 {
			return nil
		}
		return s.configService.Sets(configs)
	}
	return s.repositoryServer.UpdateByIds(ids, data)
}
