package task

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取详情
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"" validate:"numeric,min=1" label:"记录ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Task.FindById(form.Id)
	if err == nil && data != nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看任务详情", "查看计划任务["+data.Name+"]详情")
		c.SuccessWithData("获取成功", data)
	} else {
		c.Error("获取失败")
	}
}
