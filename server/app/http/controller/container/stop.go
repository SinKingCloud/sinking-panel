package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Stop 停止容器并暂停自动恢复。
func Stop(c *context.Context) {
	var form struct {
		Id string `json:"id" validate:"required,max=64" label:"容器ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Container.Stop(c.Request.Context(), form.Id); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "停止容器", "停止容器["+form.Id+"]")
		c.Success("停止成功")
	}
}
