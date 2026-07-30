package task

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Run 执行任务
func Run(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Ids {
		_ = service.Task.Run(v)
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "执行计划任务", "手动执行计划任务")
	c.Success("执行成功")
}
