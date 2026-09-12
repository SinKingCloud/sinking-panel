package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	webServer "server/app/util/server"
)

// Log 读取或清理网站访问、WAF 和进程日志。
func Log(c *context.Context) {
	var form struct {
		Id       int64  `json:"id" default:"0" validate:"required,numeric,min=1" label:"网站ID"`
		Action   string `json:"action" default:"read" validate:"required,oneof=read clear" label:"操作类型"`
		Type     string `json:"type" default:"access" validate:"required,oneof=access waf process" label:"日志类型"`
		After    int64  `json:"after" default:"0" validate:"numeric,min=0" label:"增量游标"`
		Before   int64  `json:"before" default:"0" validate:"numeric,min=0" label:"历史游标"`
		PageSize int    `json:"page_size" default:"300" validate:"numeric,min=1,max=10000" label:"读取行数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	logType := webServer.LogType(form.Type)
	typeName := map[webServer.LogType]string{
		webServer.LogAccess:  "访问",
		webServer.LogWAF:     "WAF",
		webServer.LogProcess: "进程",
	}[logType]
	if form.Action == "clear" {
		err := service.Site.ClearLog(form.Id, logType)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理网站日志", "清理网站["+strconv.FormatInt(form.Id, 10)+"]"+typeName+"日志")
			c.Success("清理成功")
		}
	} else {
		data, err := service.Site.ReadLog(form.Id, logType, form.After, form.Before, form.PageSize)
		if err != nil {
			c.Error(err.Error())
		} else {
			query := c.Request.URL.Query()
			if !query.Has("after") && !query.Has("before") {
				service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站日志", "查看网站["+strconv.FormatInt(form.Id, 10)+"]"+typeName+"日志")
			}
			c.SuccessWithData("获取成功", data)
		}
	}
}
