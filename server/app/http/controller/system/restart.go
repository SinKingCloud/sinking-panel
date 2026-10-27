package system

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Restart 重启面板或服务器。
func Restart(c *context.Context) {
	var form struct {
		Action     string `json:"action" default:"panel" validate:"required,oneof=panel server" label:"重启对象"`
		Containers bool   `json:"containers" label:"同时重启容器"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.System.Restart(form.Action, form.Containers); err != nil {
		c.Error(err.Error())
	} else {
		name := map[string]string{
			"panel":  "面板",
			"server": "服务器",
		}[form.Action]
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "重启"+name, "请求重启"+name)
		c.Success("正在重启" + name)
	}
}
