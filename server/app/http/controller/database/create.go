package database

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Create 添加数据库实例。
func Create(c *context.Context) {
	var form struct {
		Name   string `json:"name" default:"" validate:"required,max=100" label:"实例名称"`
		Type   int    `json:"type" default:"0" validate:"required" label:"数据库类型"`
		TypeId int64  `json:"type_id" default:"0" validate:"min=0" label:"分类ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &model.Database{
		Name:   form.Name,
		Type:   form.Type,
		TypeId: form.TypeId,
	}
	if err := service.Database.Create(data); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "添加数据库实例", "添加数据库实例["+strconv.FormatInt(data.Id, 10)+"]")
	c.SuccessWithData("添加成功", map[string]int64{"id": data.Id})
}
