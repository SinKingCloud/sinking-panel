package types

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
)

// Create 添加信息
func Create(c *context.Context) {
	var form struct {
		Module string `json:"module" default:"" validate:"required,max=50" label:"所属模块"`
		Name   string `json:"name" default:"" validate:"required,max=50" label:"类型名称"`
		Sort   int    `json:"sort" default:"0" validate:"numeric" label:"排序"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Type.Create(&model.Type{
		Module: form.Module,
		Name:   form.Name,
		Sort:   form.Sort,
	})
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "添加类型", "添加类型["+form.Name+"]")
		c.Success("添加成功")
	} else {
		c.Error("添加失败")
	}
}
