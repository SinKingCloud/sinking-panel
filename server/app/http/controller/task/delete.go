package task

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除信息
func Delete(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Task.Remove(form.Ids)
	if err != nil {
		c.Error("删除失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除计划任务", "删除计划任务数据")
		c.Success("删除成功")
	}
}
