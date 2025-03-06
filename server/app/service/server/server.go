package server

import (
	"server/app/util/str"
	"sync"
)

// Service 单例对象
type Service struct {
}

// obj 单例对象
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

// Server 配置表
type Server struct {
	Id         int          `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Ip         string       `gorm:"column:ip" json:"ip"`
	Port       int          `gorm:"column:port" json:"port"`
	User       string       `gorm:"column:user" json:"user"`
	AuthType   int          `gorm:"column:auth_type" json:"auth_type"`
	Name       string       `gorm:"column:name" json:"name"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
