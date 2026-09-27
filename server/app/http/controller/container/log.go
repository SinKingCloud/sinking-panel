package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Log 按游标读取或清理容器运行日志。
func Log(c *context.Context) {
	var form struct {
		Id       string `json:"id" validate:"required,max=64" label:"容器ID"`
		Action   string `json:"action" default:"read" validate:"required,oneof=read clear" label:"操作类型"`
		After    int64  `json:"after" default:"0" validate:"min=0" label:"增量游标"`
		Before   int64  `json:"before" default:"0" validate:"min=0" label:"历史游标"`
		PageSize int    `json:"page_size" default:"300" validate:"min=1,max=10000" label:"读取行数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "clear" {
		if err := service.Container.ClearLog(form.Id); err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理容器日志", "清理容器["+form.Id+"]日志")
			c.Success("清理成功")
		}
		return
	}
	data, err := service.Container.ReadLog(form.Id, form.After, form.Before, form.PageSize)
	if err != nil {
		c.Error(err.Error())
	} else {
		c.SuccessWithData("获取成功", data)
	}
}
