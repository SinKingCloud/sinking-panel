package database

import "server/app/util/str"

// SelectDatabase 数据库实例查询条件。
type SelectDatabase struct {
	TypeId  *int64
	Type    *int
	Keyword *string
	Name    *string
}

// UpdateDatabase 数据库实例更新内容，创建后不更换数据库类型。
type UpdateDatabase struct {
	TypeId *int64
	Name   *string
	Config *string
}

// Database 实例列表摘要，只返回连接是否已配置，不暴露配置内容。
type Database struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`
	Name       string       `gorm:"column:name" json:"name"`
	Type       int          `gorm:"column:type" json:"type"`
	Configured bool         `gorm:"column:configured" json:"configured"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
