package server

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
	err := service.Server.DeleteByIds(form.Ids)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除服务器", "删除服务器数据")
		c.Success("删除成功")
	} else {
		c.Error("删除失败")
	}
}
