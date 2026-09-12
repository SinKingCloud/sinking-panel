package system

import (
	"fmt"
	"strconv"

	"server/app/enum/log_type"
	repositoryLog "server/app/repository/log"
	"server/app/service"
	"server/app/util/context"
)

func Log(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,type,ip,create_time,update_time")
	var form struct {
		Action          string `json:"action" default:"list" validate:"omitempty,oneof=list clear" label:"操作类型"`
		Day             string `json:"day" default:"" validate:"omitempty,numeric,min=1" label:"保留天数"`
		Keyword         string `json:"keyword" default:"" validate:"omitempty,max=200" label:"关键词"`
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
	if form.Action == "clear" {
		day, err := strconv.Atoi(form.Day)
		if err != nil || day < 1 {
			c.Error("保留天数必须是大于0的数字")
			return
		}
		total, err := service.Log.Clear(day)
		if err != nil {
			c.Error("清理操作日志失败")
		} else {
			service.Log.Create(
				c.GetRequestIp(),
				log_type.EventDelete,
				"清理操作日志",
				fmt.Sprintf("清理%d天前的操作日志，共%d条", day, total),
			)
			c.Success(fmt.Sprintf("成功清理%d条操作日志", total))
		}
		return
	}
	where := &repositoryLog.SelectLog{}
	if form.Keyword != "" {
		where.Keyword = &form.Keyword
	}
	if form.Ip != "" {
		where.Ip = &form.Ip
	}
	if form.Location != "" {
		where.Location = &form.Location
	}
	if form.Type != "" {
		logType, err := strconv.Atoi(form.Type)
		if err != nil {
			c.Error("日志类型参数错误")
			return
		}
		where.Type = &logType
	}
	if form.Title != "" {
		where.Title = &form.Title
	}
	if form.Content != "" {
		where.Content = &form.Content
	}
	if form.CreateTimeStart != "" {
		where.CreateTimeStart = &form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where.CreateTimeEnd = &form.CreateTimeEnd
	}
	if form.UpdateTimeStart != "" {
		where.UpdateTimeStart = &form.UpdateTimeStart
	}
	if form.UpdateTimeEnd != "" {
		where.UpdateTimeEnd = &form.UpdateTimeEnd
	}
	data, err := service.Log.Select(where, query)
	if err != nil {
		c.Error("获取日志失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看操作日志", "查看操作日志列表")
		c.SuccessWithData("获取日志成功", data)
	}
}
