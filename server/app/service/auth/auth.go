package auth

import (
	serviceConfig "server/app/service/config"
	"server/app/util/cache"
)

// Service service接口
type Service interface {
	Login(account string, pwd string, types string, ip string) (string, error)
	Logout(types string) error
	UpdateAccount(account string, password string) error
	GetCaptcha(key string) (map[string]interface{}, error)
	CheckCaptcha(key string, x int, y int) bool
}

// service 注入结构
type service struct {
	configService serviceConfig.Service
	cache         cache.Interface
}

// NewService 实例化service
func NewService(configService serviceConfig.Service, cache cache.Interface) *service {
	return &service{
		configService: configService,
		cache:         cache,
	}
}
