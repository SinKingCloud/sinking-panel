package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Script 常用脚本表
type Script struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 脚本ID
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`         // 类型ID
	Name       string       `gorm:"column:name" json:"name"`               // 脚本名称
	Script     string       `gorm:"column:script" json:"script"`           // 脚本内容
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Script) TableName() string {
	return "cloud_scripts"
}

// BeforeCreate 创建前
func (s *Script) BeforeCreate(_ *gorm.DB) error {
	s.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (s *Script) BeforeUpdate(_ *gorm.DB) error {
	s.UpdateTime = str.DateTime(time.Now())
	return nil
}
