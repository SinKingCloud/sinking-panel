package database

import (
	"server/app/enum/log_type"
	repositoryDatabase "server/app/repository/database"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改数据库实例。
func Update(c *context.Context) {
	var form struct {
		Id     int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Name   string `json:"name" default:"" validate:"omitempty,max=100" label:"实例名称"`
		TypeId *int64 `json:"type_id" validate:"omitempty,min=0" label:"分类ID"`
		Config string `json:"config" default:"" validate:"omitempty,max=1048576" label:"连接配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryDatabase.UpdateDatabase{}
	if form.TypeId != nil {
		data.TypeId = form.TypeId
	}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Config != "" {
		data.Config = &form.Config
	}
	if err := service.Database.Update(form.Id, data); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改数据库实例", "修改数据库实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("修改成功")
}
