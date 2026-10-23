package database

import (
	"server/app/enum/log_type"
	repositoryDatabase "server/app/repository/database"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// List 数据库实例列表。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,name,type,create_time,update_time")
	var form struct {
		Type    string `json:"type" default:"" validate:"omitempty,numeric" label:"数据库类型"`
		TypeId  string `json:"type_id" default:"" validate:"omitempty,numeric" label:"分类ID"`
		Keyword string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		Name    string `json:"name" default:"" validate:"omitempty,max=100" label:"实例名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryDatabase.SelectDatabase{}
	if form.Type != "" {
		databaseType, err := strconv.Atoi(form.Type)
		if err != nil {
			c.Error("数据库类型参数错误")
			return
		}
		where.Type = &databaseType
	}
	if form.TypeId != "" {
		typeId, err := strconv.ParseInt(form.TypeId, 10, 64)
		if err != nil || typeId < 0 {
			c.Error("分类参数错误")
			return
		}
		where.TypeId = &typeId
	}
	if form.Keyword != "" {
		where.Keyword = &form.Keyword
	}
	if form.Name != "" {
		where.Name = &form.Name
	}
	data, err := service.Database.Select(where, query)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看数据库实例列表", "查看数据库实例列表数据")
	c.SuccessWithData("获取成功", data)
}
