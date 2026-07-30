package task

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Stop 暂停任务
func Stop(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Ids {
		_ = service.Task.Stop(v)
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "暂停计划任务", "暂停计划任务")
	c.Success("操作成功")
}
