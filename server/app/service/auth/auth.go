package auth

import (
	"server/app/service/config"
	"server/app/util/cache"
)

// Service service接口
type Service interface {
	CheckAccount(account string, pwd string) error
	GenLoginToken(types string, ip string) (string, error)
	ClearLoginToken(types string) error
	GetCaptcha(key string) (map[string]interface{}, error)
	CheckCaptcha(key string, x int, y int) bool
}

// service 注入结构
type service struct {
	configService config.Service
	cache         cache.Interface
}

// NewService 实例化service
func NewService(configService config.Service, cache cache.Interface) *service {
	return &service{
		configService: configService,
		cache:         cache,
	}
}
