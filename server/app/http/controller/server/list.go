package server

import (
	"server/app/enum/log_type"
	repositoryServer "server/app/repository/server"
	"server/app/service"
	"server/app/util/context"
)

// List 获取服务器列表
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,ip,create_time,update_time")
	var form struct {
		Ip              string `json:"ip" default:"" validate:"omitempty" label:"IP地址"`
		Port            string `json:"port" default:"" validate:"omitempty,numeric" label:"端口号"`
		User            string `json:"user" default:"" validate:"omitempty" label:"账号"`
		Name            string `json:"name" default:"" validate:"omitempty" label:"名称"`
		AuthType        string `json:"auth_type" default:"" validate:"omitempty,numeric" label:"验证方式"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
		UpdateTimeStart string `json:"update_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新起始时间"`
		UpdateTimeEnd   string `json:"update_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryServer.SelectServer{}
	if form.Ip != "" {
		where.Ip = form.Ip
	}
	if form.Port != "" {
		where.Port = form.Port
	}
	if form.User != "" {
		where.User = form.User
	}
	if form.Name != "" {
		where.Name = form.Name
	}
	if form.AuthType != "" {
		where.AuthType = form.AuthType
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
	data, err := service.Server.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看服务器列表", "查看服务器列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
