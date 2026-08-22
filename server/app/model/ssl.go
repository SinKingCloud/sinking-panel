package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Ssl SSL证书表
type Ssl struct {
	Id          int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Name        string       `gorm:"column:name" json:"name"`
	Type        int          `gorm:"column:type" json:"type"`
	Domains     []string     `gorm:"column:domains;serializer:json" json:"domains"`
	Certificate string       `gorm:"column:certificate" json:"certificate"`
	PrivateKey  string       `gorm:"column:private_key" json:"-"`
	StartTime   str.DateTime `gorm:"column:start_time" json:"start_time"`
	ExpireTime  str.DateTime `gorm:"column:expire_time" json:"expire_time"`
	CreateTime  str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime  str.DateTime `gorm:"column:update_time" json:"update_time"`
}

// TableName 获取表名
func (*Ssl) TableName() string {
	return "cloud_ssls"
}

// BeforeCreate 创建前
func (s *Ssl) BeforeCreate(_ *gorm.DB) error {
	s.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (s *Ssl) BeforeUpdate(_ *gorm.DB) error {
	s.UpdateTime = str.DateTime(time.Now())
	return nil
}
