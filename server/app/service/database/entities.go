package database

import "server/app/util/str"

// Database 返回实例详情，Config 为当前数据库类型的连接配置结构体。
type Database struct {
	Id         int64        `json:"id"`
	Name       string       `json:"name"`
	Type       int          `json:"type"`
	TypeId     int64        `json:"type_id"`
	Config     interface{}  `json:"config"`
	CreateTime str.DateTime `json:"create_time"`
	UpdateTime str.DateTime `json:"update_time"`
}
