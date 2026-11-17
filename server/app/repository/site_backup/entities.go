package site_backup

import "server/app/model"

// SelectSiteBackup 网站备份查询条件。
type SelectSiteBackup struct {
	Keyword         *string
	SiteId          *int64
	Name            *string
	SiteName        *string
	CreateTimeStart *string
	CreateTimeEnd   *string
}

// SiteBackup 网站备份列表数据，附带当前网站名称。
type SiteBackup struct {
	model.SiteBackup
	SiteName string `gorm:"column:site_name" json:"site_name"`
}
