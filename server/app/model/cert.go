package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Cert 证书表
type Cert struct {
	Id          int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`               // 证书ID
	Name        string       `gorm:"column:name" json:"name"`                       // 证书名称
	Type        int          `gorm:"column:type" json:"type"`                       // 证书类型
	Domains     []string     `gorm:"column:domains;serializer:json" json:"domains"` // 证书域名
	Certificate string       `gorm:"column:certificate" json:"certificate"`         // 证书内容
	PrivateKey  string       `gorm:"column:private_key" json:"private_key"`         // 私钥内容
	StartTime   str.DateTime `gorm:"column:start_time" json:"start_time"`           // 生效时间
	ExpireTime  str.DateTime `gorm:"column:expire_time" json:"expire_time"`         // 到期时间
	CreateTime  str.DateTime `gorm:"column:create_time" json:"create_time"`         // 创建时间
	UpdateTime  str.DateTime `gorm:"column:update_time" json:"update_time"`         // 更新时间
}

// TableName 获取表名
func (*Cert) TableName() string {
	return "cloud_certs"
}

// BeforeCreate 创建前
func (c *Cert) BeforeCreate(_ *gorm.DB) error {
	c.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (c *Cert) BeforeUpdate(_ *gorm.DB) error {
	c.UpdateTime = str.DateTime(time.Now())
	return nil
}
