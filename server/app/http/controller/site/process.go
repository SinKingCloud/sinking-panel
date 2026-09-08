package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// Process 读取、设置和控制通用网站进程。
func Process(c *context.Context) {
	var form struct {
		Action string                     `json:"action" default:"get" validate:"required,oneof=get set status start stop restart" label:"操作类型"`
		Id     int64                      `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *siteService.ProcessUpdate `json:"config" validate:"omitempty" label:"进程配置"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "status" || form.Action == "start" || form.Action == "stop" || form.Action == "restart" {
		if form.Action != "status" && c.Request.Method != "POST" {
			c.Error("进程控制请使用 POST 请求")
			return
		}
		result, err := service.Site.Process(form.Id, form.Action)
		if err != nil {
			c.Error(err.Error())
			return
		}
		message := "获取成功"
		if form.Action != "status" {
			action := map[string]string{"start": "启动", "stop": "停止", "restart": "重启"}[form.Action]
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, action+"项目进程", action+"网站["+strconv.FormatInt(form.Id, 10)+"]进程")
			message = action + "成功"
		}
		c.SuccessWithData(message, result)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetProcess(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站进程", "查看网站["+strconv.FormatInt(form.Id, 10)+"]进程配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Config == nil {
		c.Error("进程配置不能为空")
		return
	}
	if err := service.Site.UpdateProcess(form.Id, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站进程", "修改网站["+strconv.FormatInt(form.Id, 10)+"]进程配置")
	c.Success("修改成功")
}
