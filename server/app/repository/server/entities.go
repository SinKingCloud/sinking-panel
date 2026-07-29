package server

import "server/app/util/str"

// SelectServer 服务器查询条件
type SelectServer struct {
	Ip              string
	Port            string
	User            string
	Name            string
	AuthType        string
	CreateTimeStart string
	CreateTimeEnd   string
	UpdateTimeStart string
	UpdateTimeEnd   string
}

// UpdateServer 服务器更新
type UpdateServer struct {
	Ip       interface{}
	Port     interface{}
	User     interface{}
	AuthType interface{}
	Password interface{}
	Name     interface{}
}

// Server 配置表
type Server struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Ip         string       `gorm:"column:ip" json:"ip"`
	Port       int          `gorm:"column:port" json:"port"`
	User       string       `gorm:"column:user" json:"user"`
	AuthType   int          `gorm:"column:auth_type" json:"auth_type"`
	Name       string       `gorm:"column:name" json:"name"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
