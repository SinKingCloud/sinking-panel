package task

import (
	"server/app/service"
	"server/app/util/context"
)

// Run 执行任务
func Run(c *context.Context) {
	type Form struct {
		Ids []int `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Ids {
		_ = service.Task.Run(v)
	}
	c.Success("执行成功")
}
