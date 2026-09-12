package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// Secret 云服务及其他服务的密钥凭据。
type Secret struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 凭据ID
	Name       string       `gorm:"column:name" json:"name"`               // 凭据名称
	Provider   int          `gorm:"column:provider" json:"provider"`       // 服务商，参见 secret_provider
	Data       string       `gorm:"column:data" json:"data"`               // 凭据内容的 JSON 对象字符串
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名。
func (*Secret) TableName() string {
	return "cloud_secrets"
}

// BeforeCreate 创建前。
func (s *Secret) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	s.CreateTime = now
	s.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前。
func (s *Secret) BeforeUpdate(_ *gorm.DB) error {
	s.UpdateTime = str.DateTime(time.Now())
	return nil
}
