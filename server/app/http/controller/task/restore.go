package task

import (
	"server/app/service"
	"server/app/util/context"
)

// Restore 恢复任务
func Restore(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Ids {
		_ = service.Task.Restore(v)
	}
	c.Success("操作成功")
}
