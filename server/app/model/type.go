package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Type 类型表
type Type struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Module     string       `gorm:"column:module" json:"module"`
	Name       string       `gorm:"column:name" json:"name"`
	Sort       int          `gorm:"column:sort" json:"sort"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}

// TableName 获取表名
func (*Type) TableName() string {
	return "cloud_types"
}

// BeforeCreate 创建前
func (t *Type) BeforeCreate(_ *gorm.DB) error {
	t.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (t *Type) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
