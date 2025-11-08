package task

import (
	"server/app/service"
	"server/app/util/context"
)

// Stop 暂停任务
func Stop(c *context.Context) {
	type Form struct {
		Ids []int `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Ids {
		_ = service.Task.Stop(v)
	}
	c.Success("操作成功")
}
