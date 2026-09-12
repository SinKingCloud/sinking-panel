package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Process 读取、设置和控制通用网站进程。
func Process(c *context.Context) {
	var form struct {
		Action string                     `json:"action" default:"get" validate:"required,oneof=get set status start stop restart" label:"操作类型"`
		Id     int64                      `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.ProcessUpdate `json:"config" validate:"omitempty" label:"进程配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "status" || form.Action == "start" || form.Action == "stop" || form.Action == "restart" {
		if form.Action != "status" && c.Request.Method != "POST" {
			c.Error("进程控制请使用 POST 请求")
			return
		}
		data, err := service.Site.Process(form.Id, form.Action)
		if err != nil {
			c.Error(err.Error())
		} else {
			msg := "获取成功"
			if form.Action != "status" {
				action := map[string]string{"start": "启动", "stop": "停止", "restart": "重启"}[form.Action]
				service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, action+"项目进程", action+"网站["+strconv.FormatInt(form.Id, 10)+"]进程")
				msg = action + "成功"
			}
			c.SuccessWithData(msg, data)
		}
	} else if form.Action == "get" {
		data, err := service.Site.GetProcess(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站进程", "查看网站["+strconv.FormatInt(form.Id, 10)+"]进程配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		if form.Config == nil {
			c.Error("进程配置不能为空")
			return
		}
		err := service.Site.UpdateProcess(form.Id, form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站进程", "修改网站["+strconv.FormatInt(form.Id, 10)+"]进程配置")
			c.Success("修改成功")
		}
	}
}
