package config

import (
	"server/app/repository/config"
	"server/app/util/cache"
)

// Service service接口
type Service interface {
	Group(group string) map[string]string
	Set(key string, value string) error
	Sets(configs map[string]string) error
	Get(group string, key string) string
	Load(group string, key string) (string, error)
}

// service 注入结构
type service struct {
	repositoryConfig config.Interface
	cache            cache.Interface
}

// NewService 实例化service
func NewService(repositoryConfig config.Interface, cache cache.Interface) *service {
	return &service{
		repositoryConfig: repositoryConfig,
		cache:            cache,
	}
}
