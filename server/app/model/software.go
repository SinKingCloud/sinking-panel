package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Software 软件包及其安装信息，不保存容器的实时运行状态。
type Software struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 软件 ID
	Name       string       `gorm:"column:name" json:"name"`               // 显示名称
	Code       string       `gorm:"column:code" json:"code"`               // 软件唯一标识
	Version    string       `gorm:"column:version" json:"version"`         // 当前软件包版本
	Status     int          `gorm:"column:status" json:"status"`           // 安装状态
	Config     string       `gorm:"column:config" json:"config"`           // 插件内部配置
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名。
func (*Software) TableName() string {
	return "cloud_software"
}

// BeforeCreate 创建前。
func (s *Software) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	s.CreateTime = now
	s.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前。
func (s *Software) BeforeUpdate(_ *gorm.DB) error {
	s.UpdateTime = str.DateTime(time.Now())
	return nil
}
