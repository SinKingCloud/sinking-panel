package script

import "server/app/util/str"

// SelectScript 常用脚本查询条件
type SelectScript struct {
	TypeId  string
	Keyword string
	Name    string
	Script  string
}

// UpdateScript 常用脚本更新
type UpdateScript struct {
	TypeId interface{}
	Name   interface{}
	Script interface{}
}

// Script 常用脚本列表数据
type Script struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`
	Name       string       `gorm:"column:name" json:"name"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
