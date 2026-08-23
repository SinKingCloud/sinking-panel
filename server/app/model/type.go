package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Type 类型表
type Type struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 类型ID
	Module     string       `gorm:"column:module" json:"module"`           // 所属模块
	Name       string       `gorm:"column:name" json:"name"`               // 类型名称
	Sort       int64        `gorm:"column:sort" json:"sort"`               // 排序值
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
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
