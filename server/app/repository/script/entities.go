package script

import "server/app/util/str"

// SelectScript 常用脚本查询条件
type SelectScript struct {
	Module  *string
	TypeId  *int64
	Keyword *string
	Name    *string
	Script  *string
}

// UpdateScript 常用脚本更新
type UpdateScript struct {
	Name   *string
	Script *string
}

// Script 常用脚本列表数据
type Script struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Module     string       `gorm:"column:module" json:"module"`
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`
	Name       string       `gorm:"column:name" json:"name"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
