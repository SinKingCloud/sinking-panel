package container

import (
	"strings"

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
