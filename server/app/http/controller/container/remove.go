package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strings"
)

// Remove 批量停止并删除容器，保留挂载的业务数据。
func Remove(c *context.Context) {
	var form struct {
		Ids []string `json:"ids" default:"" validate:"required,min=1,max=1000,unique,dive,required,max=64" label:"容器ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Container.Remove(form.Ids); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除容器", "删除容器["+strings.Join(form.Ids, ",")+"]，保留挂载数据")
		c.Success("删除成功")
	}
}
