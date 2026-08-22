package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Site 网站表
type Site struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Name       string       `gorm:"column:name" json:"name"`
	Type       int          `gorm:"column:type" json:"type"`
	Status     int          `gorm:"column:status" json:"status"`
	Root       string       `gorm:"column:root" json:"root"`
	RunPath    string       `gorm:"column:run_path" json:"run_path"`
	SslIds     []int64      `gorm:"column:ssl_ids;serializer:json" json:"ssl_ids"`
	Ssl        string       `gorm:"column:ssl" json:"ssl"`
	Limit      string       `gorm:"column:limit" json:"limit"`
	Waf        string       `gorm:"column:waf" json:"waf"`
	Cache      string       `gorm:"column:cache" json:"cache"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}

// TableName 获取表名
func (*Site) TableName() string {
	return "cloud_sites"
}

// BeforeCreate 创建前
func (s *Site) BeforeCreate(_ *gorm.DB) error {
	s.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (s *Site) BeforeUpdate(_ *gorm.DB) error {
	s.UpdateTime = str.DateTime(time.Now())
	return nil
}
