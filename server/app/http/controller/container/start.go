package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Start 启动已有容器。
func Start(c *context.Context) {
	var form struct {
		Id string `json:"id" validate:"required,max=64" label:"容器ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Container.Start(c.Request.Context(), form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "启动容器", "启动容器["+form.Id+"]")
		c.SuccessWithData("启动成功", data)
	}
}
