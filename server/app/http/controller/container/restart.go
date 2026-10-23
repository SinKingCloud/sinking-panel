package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Restart 重启容器并应用受管软件的待生效配置。
func Restart(c *context.Context) {
	var form struct {
		Id string `json:"id" validate:"required,max=64" label:"容器ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Container.Restart(c.Request.Context(), form.Id); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "重启容器", "重启容器["+form.Id+"]")
		c.Success("重启成功")
	}
}
