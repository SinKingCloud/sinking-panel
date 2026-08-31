package site

import "server/app/util/str"

// SelectSite 网站查询条件。
type SelectSite struct {
	Keyword         string
	Name            string
	TypeId          string
	Type            string
	Status          string
	CreateTimeStart string
	CreateTimeEnd   string
	UpdateTimeStart string
	UpdateTimeEnd   string
}

// UpdateSite 网站更新内容。
type UpdateSite struct {
	Name    *string
	TypeId  *int64
	Type    *int
	Status  *int
	Root    *string
	RunPath *string
	Config  *string
}

// Site 网站列表数据，不在列表中返回完整配置。
type Site struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`
	Name       string       `gorm:"column:name" json:"name"`
	Type       int          `gorm:"column:type" json:"type"`
	Status     int          `gorm:"column:status" json:"status"`
	Root       string       `gorm:"column:root" json:"root"`
	RunPath    string       `gorm:"column:run_path" json:"run_path"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
