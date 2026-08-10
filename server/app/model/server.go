package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Server 配置表
type Server struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Ip         string       `gorm:"column:ip" json:"ip"`
	Port       int          `gorm:"column:port" json:"port"`
	User       string       `gorm:"column:user" json:"user"`
	AuthType   int          `gorm:"column:auth_type" json:"auth_type"`
	Password   string       `gorm:"column:password" json:"-"`
	Name       string       `gorm:"column:name" json:"name"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}

// TableName 获取表名
func (*Server) TableName() string {
	return "cloud_servers"
}

// BeforeCreate 创建前
func (t *Server) BeforeCreate(_ *gorm.DB) error {
	t.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (t *Server) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
