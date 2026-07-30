package system

import (
	repositoryLog "server/app/repository/log"
	"server/app/service"
	"server/app/util/context"
)

func Log(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,type,ip,create_time,update_time")
	var form struct {
		Type            string `json:"type" default:"" validate:"omitempty,numeric" label:"类型"`
		Ip              string `json:"ip" default:"" validate:"omitempty" label:"IP地址"`
		Location        string `json:"location" default:"" validate:"omitempty,max=100" label:"IP归属地"`
		Title           string `json:"title" default:"" validate:"omitempty" label:"标题"`
		Content         string `json:"content" default:"" validate:"omitempty" label:"内容"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
		UpdateTimeStart string `json:"update_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新起始时间"`
		UpdateTimeEnd   string `json:"update_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryLog.SelectLog{}
	if form.Ip != "" {
		where.Ip = form.Ip
	}
	if form.Location != "" {
		where.Location = form.Location
	}
	if form.Type != "" {
		where.Type = form.Type
	}
	if form.Title != "" {
		where.Title = form.Title
	}
	if form.Content != "" {
		where.Content = form.Content
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
	data, err := service.Log.Select(where, query)
	if err != nil {
		c.Error("获取日志失败")
	} else {
		c.SuccessWithData("获取日志成功", data)
	}
}
