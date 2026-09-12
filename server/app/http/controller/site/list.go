package site

import (
	"server/app/enum/log_type"
	repositorySite "server/app/repository/site"
	"server/app/service"
	"server/app/util/context"
)

// List 获取网站列表。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,type_id,name,type,status,create_time,update_time")
	var form struct {
		Keyword         string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		Name            string `json:"name" default:"" validate:"omitempty,max=100" label:"网站名称"`
		TypeId          string `json:"type_id" default:"" validate:"omitempty,numeric" label:"网站分类ID"`
		Type            string `json:"type" default:"" validate:"omitempty,numeric" label:"网站类型"`
		Status          string `json:"status" default:"" validate:"omitempty,numeric" label:"网站状态"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
		UpdateTimeStart string `json:"update_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新起始时间"`
		UpdateTimeEnd   string `json:"update_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositorySite.SelectSite{}
	if form.Keyword != "" {
		where.Keyword = form.Keyword
	}
	if form.Name != "" {
		where.Name = form.Name
	}
	if form.TypeId != "" {
		where.TypeId = form.TypeId
	}
	if form.Type != "" {
		where.Type = form.Type
	}
	if form.Status != "" {
		where.Status = form.Status
	}
	if form.CreateTimeStart != "" {
		where.CreateTimeStart = form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where.CreateTimeEnd = form.CreateTimeEnd
	}
	if form.UpdateTimeStart != "" {
		where.UpdateTimeStart = form.UpdateTimeStart
	}
	if form.UpdateTimeEnd != "" {
		where.UpdateTimeEnd = form.UpdateTimeEnd
	}
	data, err := service.Site.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站列表", "查看网站列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
