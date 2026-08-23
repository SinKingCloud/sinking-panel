package model

import (
	"gorm.io/gorm"
	"server/app/util/str"
	"time"
)

// Log 日志表
type Log struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 日志ID
	Type       int          `gorm:"column:type" json:"type"`               // 日志类型
	Ip         string       `gorm:"column:ip" json:"ip"`                   // 操作IP
	Location   string       `gorm:"column:location" json:"location"`       // IP归属地
	Title      string       `gorm:"column:title" json:"title"`             // 日志标题
	Content    string       `gorm:"column:content" json:"content"`         // 日志内容
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Log) TableName() string {
	return "cloud_logs"
}

// BeforeCreate 创建前
func (t *Log) BeforeCreate(_ *gorm.DB) error {
	t.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (t *Log) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
