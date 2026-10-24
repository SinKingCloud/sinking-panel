package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Database 数据库实例连接信息。
type Database struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 实例ID
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`         // 分类ID
	Name       string       `gorm:"column:name" json:"name"`               // 实例名称
	Type       int          `gorm:"column:type" json:"type"`               // 数据库类型，参见 database_kind
	Config     string       `gorm:"column:config" json:"-"`                // 连接配置，由 service 加密保存
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名。
func (*Database) TableName() string {
	return "cloud_databases"
}

// BeforeCreate 创建前。
func (d *Database) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	d.CreateTime = now
	d.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前。
func (d *Database) BeforeUpdate(_ *gorm.DB) error {
	d.UpdateTime = str.DateTime(time.Now())
	return nil
}
