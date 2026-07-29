package task

import (
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
	for _, v := range form.Ids {
		_ = service.Task.Remove(v)
	}
	c.Success("删除成功")
}
