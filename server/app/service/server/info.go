package server

import (
	"errors"
	"strconv"
	"strings"

	"server/app/constant"
	"server/app/enum/server_auth_type"
	"server/app/model"
)

// FindById 查询信息并解密凭据。
func (s *service) FindById(id int64) (*model.Server, error) {
	if id < 0 {
		return nil, errors.New("服务器ID不合法")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var data *model.Server
	var err error
	if id == 0 {
		configs := s.configService.Group(constant.SshGroup)
		ip := strings.TrimSpace(configs[constant.SshIP])
		if ip == "" {
			ip = "127.0.0.1"
		}
		port, err := strconv.Atoi(configs[constant.SshPort])
		if err != nil {
			port = 22
		}
		authType, err := strconv.Atoi(configs[constant.SshAuthType])
		if err != nil {
			authType = server_auth_type.Password
		}
		name := strings.TrimSpace(configs[constant.SshName])
		if name == "" {
			name = "本机终端"
		}
		data = &model.Server{
			Id:       0,
			Ip:       ip,
			Port:     port,
			User:     strings.TrimSpace(configs[constant.SshUser]),
			AuthType: authType,
			Password: configs[constant.SshPassword],
			Name:     name,
		}
	} else {
		data, err = s.repositoryServer.FindById(id)
		if err != nil || data == nil {
			return nil, err
		}
	}
	data.Password = s.decrypt(data.Ip, data.Password)
	return data, nil
}
