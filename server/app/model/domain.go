package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Domain 网站域名表
type Domain struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 域名ID
	SiteId     int64        `gorm:"column:site_id" json:"site_id"`         // 所属网站ID
	CertId     int64        `gorm:"column:cert_id" json:"cert_id"`         // 证书ID，0表示未开启SSL
	Domain     string       `gorm:"column:domain" json:"domain"`           // 域名
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Domain) TableName() string {
	return "cloud_domains"
}

// BeforeCreate 创建前
func (d *Domain) BeforeCreate(_ *gorm.DB) error {
	d.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (d *Domain) BeforeUpdate(_ *gorm.DB) error {
	d.UpdateTime = str.DateTime(time.Now())
	return nil
}
