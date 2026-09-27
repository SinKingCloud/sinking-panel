package redis

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Log 查看或清理 Redis 运行日志。
func Log(c *context.Context) {
	var form struct {
		Action   string `json:"action" default:"read" validate:"required,oneof=read clear" label:"操作类型"`
		After    int64  `json:"after" default:"0" validate:"numeric,min=0" label:"增量游标"`
		Before   int64  `json:"before" default:"0" validate:"numeric,min=0" label:"历史游标"`
		PageSize int    `json:"page_size" default:"300" validate:"numeric,min=1,max=10000" label:"读取行数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "clear" {
		if err := service.Container.ClearLog(service.Software.Redis().InstanceID); err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理 Redis 日志", "清理 Redis 运行日志")
			c.Success("清理成功")
		}
	} else {
		data, err := service.Container.ReadLog(service.Software.Redis().InstanceID, form.After, form.Before, form.PageSize)
		if err != nil {
			c.Error(err.Error())
		} else {
			query := c.Request.URL.Query()
			if !query.Has("after") && !query.Has("before") {
				service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看 Redis 日志", "查看 Redis 运行日志")
			}
			c.SuccessWithData("获取成功", data)
		}
	}
}
