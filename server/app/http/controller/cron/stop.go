package cron

import (
	"server/app/service"
	"server/app/util/server"
)

// Stop 暂停任务
func Stop(c *server.Context) {
	type Form struct {
		Ids []int `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Ids {
		_ = service.Cron.Stop(v)
	}
	c.Success("操作成功")
}
