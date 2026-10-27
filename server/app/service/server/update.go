package server

import (
	"errors"
	"strconv"
	"strings"

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
	s.mu.Lock()
	defer s.mu.Unlock()

	// 本机连接保存在 SSH 配置组中。
	if len(ids) == 1 && ids[0] == 0 {
		configs := make(map[string]string)
		if data.Ip != nil {
			configs[constant.SshIP] = strings.TrimSpace(*data.Ip)
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
		if data.Name != nil {
			configs[constant.SshName] = *data.Name
		}
		if len(configs) == 0 && data.Password == nil {
			return nil
		}
		stored := s.configService.Group(constant.SshGroup)
		oldIP := strings.TrimSpace(stored[constant.SshIP])
		if oldIP == "" {
			oldIP = "127.0.0.1"
		}
		ip := oldIP
		if data.Ip != nil {
			ip = configs[constant.SshIP]
		}
		if ip == "" {
			ip = "127.0.0.1"
		}
		if data.Password != nil || ip != oldIP {
			password := stored[constant.SshPassword]
			if data.Password != nil {
				password = *data.Password
			} else {
				password = s.decrypt(oldIP, password)
			}
			configs[constant.SshPassword] = s.encrypt(ip, password)
		}
		return s.configService.Sets(configs)
	}

	// 普通字段或同时指定 IP、密码时，直接批量更新。
	if data.Ip == nil && data.Password == nil {
		return s.repositoryServer.UpdateByIds(ids, data)
	}
	if data.Password != nil && data.Ip != nil {
		password := s.encrypt(*data.Ip, *data.Password)
		updated := *data
		updated.Password = &password
		return s.repositoryServer.UpdateByIds(ids, &updated)
	}

	// 需要旧凭据时，先组装各记录的更新数据，再写入。
	records, err := s.repositoryServer.SelectIdDataMapByIds(ids)
	if err != nil {
		return err
	}
	updates := make(map[int64]*repositoryServer.UpdateServer, len(records))
	batchIds := make([]int64, 0, len(ids))
	for _, id := range ids {
		current := records[id]
		if current == nil {
			return errors.New("服务器不存在")
		}
		ip := current.Ip
		if data.Ip != nil {
			ip = *data.Ip
		}
		if data.Password == nil && current.Ip == ip {
			batchIds = append(batchIds, id)
			continue
		}
		password := current.Password
		if data.Password != nil {
			password = *data.Password
		} else {
			password = s.decrypt(current.Ip, password)
		}
		password = s.encrypt(ip, password)
		value := *data
		value.Password = &password
		updates[id] = &value
	}
	for _, id := range ids {
		if value := updates[id]; value != nil {
			if err := s.repositoryServer.UpdateByIds([]int64{id}, value); err != nil {
				return err
			}
		}
	}
	return s.repositoryServer.UpdateByIds(batchIds, data)
}
