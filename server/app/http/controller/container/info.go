package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 查询容器的完整配置和运行信息。
func Info(c *context.Context) {
	var form struct {
		Id string `json:"id" validate:"required,max=64" label:"容器ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Container.Info(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看容器详情", "查看容器["+form.Id+"]详情")
		c.SuccessWithData("获取成功", data)
	}
}
