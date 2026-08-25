package system

import (
	"net/http"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// HTTP 控制网站 HTTP 服务并查询运行状态。
func HTTP(c *context.Context) {
	var form struct {
		Action    string                  `json:"action" default:"status" validate:"required,oneof=status start stop restart sync get set log" label:"操作类型"`
		LogAction string                  `json:"log_action" default:"read" validate:"required,oneof=read clear" label:"日志操作"`
		After     int64                   `json:"after" default:"0" validate:"numeric,min=0" label:"增量游标"`
		Before    int64                   `json:"before" default:"0" validate:"numeric,min=0" label:"历史游标"`
		PageSize  int                     `json:"page_size" default:"300" validate:"numeric,min=1,max=10000" label:"读取行数"`
		Config    *siteService.HTTPUpdate `json:"config" validate:"omitempty" label:"HTTP配置"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	var err error
	switch form.Action {
	case "get":
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站服务", "查看网站 HTTP 配置")
		c.SuccessWithData("获取成功", sinking_web.H{"running": service.Site.Running(), "config": service.Site.GetHTTP()})
		return
	case "log":
		if form.LogAction == "clear" {
			if c.Request.Method != http.MethodPost {
				c.Error("清理日志仅支持 POST 请求")
				return
			}
			if err = service.Site.ClearServerLog(); err != nil {
				c.Error(err.Error())
				return
			}
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理服务日志", "清理网站 HTTP 服务运行日志")
			c.Success("清理成功")
			return
		}
		result, readErr := service.Site.ReadServerLog(form.After, form.Before, form.PageSize)
		if readErr != nil {
			c.Error(readErr.Error())
			return
		}
		query := c.Request.URL.Query()
		if !query.Has("after") && !query.Has("before") {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看服务日志", "查看网站 HTTP 服务运行日志")
		}
		c.SuccessWithData("获取成功", result)
		return
	case "set":
		if form.Config == nil {
			c.Error("HTTP配置不能为空")
			return
		}
		err = service.Site.UpdateHTTP(form.Config)
	case "start":
		err = service.Site.Start()
	case "stop":
		err = service.Site.Stop()
	case "restart":
		err = service.Site.Restart()
	case "sync":
		err = service.Site.Sync()
	case "status":
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站服务", "查看网站 HTTP 服务运行状态")
		c.SuccessWithData("获取成功", sinking_web.H{"running": service.Site.Running()})
		return
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	name := map[string]string{"set": "修改", "start": "启动", "stop": "停止", "restart": "重启", "sync": "同步"}[form.Action]
	detail := name + "网站 HTTP 服务"
	if form.Action == "set" {
		detail = "修改网站 HTTP 配置"
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, name+"网站服务", detail)
	c.SuccessWithData(name+"成功", sinking_web.H{"running": service.Site.Running()})
}
