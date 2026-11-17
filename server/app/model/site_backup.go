package model

import (
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// SiteBackup 网站备份记录，每次归档对应一条记录。
type SiteBackup struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 备份ID
	SiteId     int64        `gorm:"column:site_id" json:"site_id"`         // 网站ID
	Name       string       `gorm:"column:name" json:"name"`               // 备份文件名
	Size       int64        `gorm:"column:size" json:"size"`               // 归档大小，单位字节
	Data       string       `gorm:"column:data" json:"data"`               // 备份副本及归档信息的 JSON 字符串
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
}

// TableName 获取表名。
func (*SiteBackup) TableName() string {
	return "cloud_site_backups"
}

// BeforeCreate 创建前。
func (s *SiteBackup) BeforeCreate(_ *gorm.DB) error {
	s.CreateTime = str.DateTime(time.Now())
	return nil
}
