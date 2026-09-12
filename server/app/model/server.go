package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Server 服务器表
type Server struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 服务器ID
	Ip         string       `gorm:"column:ip" json:"ip"`                   // 服务器IP
	Port       int          `gorm:"column:port" json:"port"`               // SSH端口
	User       string       `gorm:"column:user" json:"user"`               // 登录用户
	AuthType   int          `gorm:"column:auth_type" json:"auth_type"`     // 认证类型
	Password   string       `gorm:"column:password" json:"-"`              // 登录密码
	Name       string       `gorm:"column:name" json:"name"`               // 服务器名称
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Server) TableName() string {
	return "cloud_servers"
}

// BeforeCreate 创建前
func (t *Server) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	t.CreateTime = now
	t.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前
func (t *Server) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
