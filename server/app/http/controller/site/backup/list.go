package backup

import (
	"server/app/enum/log_type"
	repositorySiteBackup "server/app/repository/site_backup"
	"server/app/service"
	"server/app/util/context"
)

// List 获取网站备份列表。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,site_id,name,size,create_time")
	var form struct {
		Keyword         string `json:"keyword" default:"" validate:"omitempty,max=255" label:"搜索关键词"`
		SiteId          int64  `json:"site_id" default:"0" validate:"min=0" label:"网站ID"`
		Name            string `json:"name" default:"" validate:"omitempty,max=255" label:"备份名称"`
		SiteName        string `json:"site_name" default:"" validate:"omitempty,max=255" label:"网站名称"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositorySiteBackup.SelectSiteBackup{}
	if form.Keyword != "" {
		where.Keyword = &form.Keyword
	}
	if form.SiteId > 0 {
		where.SiteId = &form.SiteId
	}
	if form.Name != "" {
		where.Name = &form.Name
	}
	if form.SiteName != "" {
		where.SiteName = &form.SiteName
	}
	if form.CreateTimeStart != "" {
		where.CreateTimeStart = &form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where.CreateTimeEnd = &form.CreateTimeEnd
	}
	data, err := service.Site.SelectBackup(where, query)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站备份", "查看网站备份列表")
		c.SuccessWithData("获取成功", data)
	}
}
