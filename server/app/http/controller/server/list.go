package server

import (
	"server/app/service"
	"server/app/util/page"
	"server/app/util/server"
)

// List 获取服务器列表
func List(c *server.Context) {
	pageInfo := page.ValidatePageDefault(c)
	type Form struct {
		OrderByField    string `json:"order_by_field" default:"sort" validate:"oneof=id ip create_time update_time" label:"排序字段"`
		OrderByType     string `json:"order_by_type" default:"desc" validate:"oneof=desc asc" label:"排序类型"`
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
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	where := make(map[string]string)
	if form.Ip != "" {
		where["ip"] = form.Ip
	}
	if form.Port != "" {
		where["port"] = form.Port
	}
	if form.User != "" {
		where["user"] = form.User
	}
	if form.Name != "" {
		where["name"] = form.Name
	}
	if form.AuthType != "" {
		where["auth_type"] = form.AuthType
	}
	if form.CreateTimeStart != "" {
		where["create_time_start"] = form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where["create_time_end"] = form.CreateTimeEnd
	}
	if form.UpdateTimeStart != "" {
		where["update_time_start"] = form.UpdateTimeStart
	}
	if form.UpdateTimeEnd != "" {
		where["update_time_end"] = form.UpdateTimeEnd
	}
	data, total, err := service.Server.Select(where, form.OrderByField, form.OrderByType, pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", page.NewPage(total, pageInfo.Page, pageInfo.PageSize, data))
	}
}
