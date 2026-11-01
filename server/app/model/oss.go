package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Oss 对象存储连接信息，配置以 JSON 保存，凭据由 SecretId 关联的密钥记录提供。
type Oss struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 存储ID
	Name       string       `gorm:"column:name" json:"name"`               // 存储名称
	SecretId   int64        `gorm:"column:secret_id" json:"secret_id"`     // 密钥ID
	Config     string       `gorm:"column:config" json:"config"`           // 存储配置的 JSON 对象字符串
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名。
func (*Oss) TableName() string {
	return "cloud_oss"
}

// BeforeCreate 创建前设置时间。
func (o *Oss) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	o.CreateTime = now
	o.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前设置时间。
func (o *Oss) BeforeUpdate(_ *gorm.DB) error {
	o.UpdateTime = str.DateTime(time.Now())
	return nil
}
