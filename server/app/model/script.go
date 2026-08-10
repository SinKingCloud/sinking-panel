package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Script 常用脚本表
type Script struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`
	Name       string       `gorm:"column:name" json:"name"`
	Script     string       `gorm:"column:script" json:"script"`
	Sort       int          `gorm:"column:sort" json:"sort"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
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
