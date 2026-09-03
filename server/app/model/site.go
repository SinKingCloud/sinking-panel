package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Site 网站表
type Site struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 网站ID
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`         // 网站分类ID
	Name       string       `gorm:"column:name" json:"name"`               // 网站名称
	Type       int          `gorm:"column:type" json:"type"`               // 网站类型
	Status     int          `gorm:"column:status" json:"status"`           // 网站状态
	Root       string       `gorm:"column:root" json:"root"`               // 网站根目录
	RunPath    string       `gorm:"column:run_path" json:"run_path"`       // 网站运行目录
	Config     string       `gorm:"column:config" json:"config"`           // 网站配置
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Site) TableName() string {
	return "cloud_sites"
}

// BeforeCreate 创建前
func (s *Site) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	s.CreateTime = now
	s.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前
func (s *Site) BeforeUpdate(_ *gorm.DB) error {
	s.UpdateTime = str.DateTime(time.Now())
	return nil
}
