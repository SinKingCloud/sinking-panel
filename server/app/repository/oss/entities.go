package oss

import (
	"server/app/util/str"
)

// SelectOss 对象存储查询条件。
type SelectOss struct {
	Keyword  *string
	SecretId *int64
}

// UpdateOss 对象存储更新字段，关联密钥创建后不可修改。
type UpdateOss struct {
	Name   *string
	Config *string
}

// Oss 对象存储详情和列表，不包含密钥正文。
type Oss struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 存储ID
	Name       string       `gorm:"column:name" json:"name"`               // 存储名称
	SecretId   int64        `gorm:"column:secret_id" json:"secret_id"`     // 密钥ID
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}
