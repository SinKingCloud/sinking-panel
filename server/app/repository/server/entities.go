package server

import "server/app/util/str"

// SelectServer 服务器查询条件
type SelectServer struct {
	Keyword         *string
	Ip              *string
	Port            *int
	User            *string
	Name            *string
	AuthType        *int
	CreateTimeStart *string
	CreateTimeEnd   *string
	UpdateTimeStart *string
	UpdateTimeEnd   *string
}

// UpdateServer 服务器更新
type UpdateServer struct {
	Ip       *string
	Port     *int
	User     *string
	AuthType *int
	Password *string
	Name     *string
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
