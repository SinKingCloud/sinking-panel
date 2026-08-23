package model

import (
	"gorm.io/gorm"
	"server/app/util/str"
	"time"
)

// Config 配置表
type Config struct {
	Key        string       `gorm:"column:key" json:"key"`                 // 配置键
	Value      string       `gorm:"column:value" json:"value"`             // 配置值
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Config) TableName() string {
	return "cloud_configs"
}

// BeforeCreate 创建前
func (t *Config) BeforeCreate(_ *gorm.DB) error {
	t.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (t *Config) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
