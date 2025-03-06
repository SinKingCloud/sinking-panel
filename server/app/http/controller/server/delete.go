package server

import (
	"server/app/service"
	"server/app/util/server"
)

// Delete 删除信息
func Delete(c *server.Context) {
	type Form struct {
		Ids []int `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	err := service.Server.DeleteByIds(form.Ids)
	if err == nil {
		c.Success("删除成功")
	} else {
		c.Error("删除失败")
	}
}
