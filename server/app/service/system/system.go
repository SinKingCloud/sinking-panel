package system

import (
	"sync"
)

// Service 系统服务
type Service struct {
}

// 单例对象
var (
	obj  *Service
	once sync.Once
)

// GetIns 获取单例
func GetIns() *Service {
	once.Do(func() {
		obj = &Service{}
	})
	return obj
}
