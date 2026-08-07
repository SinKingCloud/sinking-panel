package task

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

func Log(c *context.Context) {
	var form struct {
		Id       int64  `json:"id" default:"" validate:"required,numeric,min=1" label:"记录ID"`
		Action   string `json:"action" default:"read" validate:"omitempty,oneof=read clear" label:"操作类型"`
		Page     int    `json:"page" default:"1" validate:"numeric,min=1" label:"分页页码"`
		PageSize int    `json:"page_size" default:"10000" validate:"numeric,min=1,max=10000" label:"分页容量"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Task.FindById(form.Id)
	if err != nil || data == nil {
		c.Error("获取失败")
		return
	}
	if form.Action == "clear" {
		if err = service.Task.ClearLog(data.Id); err != nil {
			c.Error("清理日志失败")
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理任务日志", "清理任务日志")
		c.Success("清理成功")
		return
	}
	logs := service.Task.ReadLog(data.Id, form.Page, form.PageSize)
	c.SuccessWithData("获取成功", logs)
}
