package script

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
)

// Create 添加常用脚本
func Create(c *context.Context) {
	var form struct {
		TypeId int64  `json:"type_id" default:"0" validate:"min=0" label:"类型ID"`
		Name   string `json:"name" default:"" validate:"required,max=100" label:"脚本名称"`
		Script string `json:"script" default:"" validate:"required,max=131072" label:"脚本内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Script.Create(&model.Script{
		TypeId: form.TypeId,
		Name:   form.Name,
		Script: form.Script,
	}); err != nil {
		c.Error("添加失败")
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "添加常用脚本", "添加常用脚本["+form.Name+"]")
	c.Success("添加成功")
}
