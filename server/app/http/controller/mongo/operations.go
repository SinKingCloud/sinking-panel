package mongo

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Databases 管理数据库，系统数据库由业务层保护。
func Databases(c *context.Context) {
	m := service.Software.Mongo()
	var form struct {
		Action string `json:"action" default:"list" validate:"oneof=list create delete" label:"操作"`
		Name   string `json:"name" default:"" label:"数据库名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "list" {
		list, err := m.Databases(c.Request.Context())
		if err != nil {
			c.Error(err.Error())
			return
		}
		c.SuccessWithData("获取成功", map[string]interface{}{"list": list, "total": len(list)})
		return
	}
	if err := m.ChangeDatabase(c.Request.Context(), form.Action, form.Name); err != nil {
		c.Error(err.Error())
		return
	}
	title, kind := "创建数据库", log_type.EventCreate
	if form.Action == "delete" {
		title, kind = "删除数据库", log_type.EventDelete
	}
	service.Log.Create(c.GetRequestIp(), kind, title, m.InstanceID+" / "+form.Name)
	c.Success("操作成功")
}

// Users 管理业务账号，审计记录不包含密码。
func Users(c *context.Context) {
	m := service.Software.Mongo()
	var form struct {
		Action   string `json:"action" default:"list" validate:"oneof=list create delete password" label:"操作"`
		Name     string `json:"name" default:"" label:"账号"`
		Password string `json:"password" default:"" label:"密码"`
		Database string `json:"database" default:"" label:"关联数据库"`
		Host     string `json:"host" default:"" label:"来源"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "list" {
		list, err := m.Users(c.Request.Context())
		if err != nil {
			c.Error(err.Error())
			return
		}
		c.SuccessWithData("获取成功", map[string]interface{}{"list": list, "total": len(list)})
		return
	}
	if err := m.ChangeUser(c.Request.Context(), form.Action, form.Name, form.Password, form.Database, form.Host); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "管理数据库账号", m.InstanceID+" / "+form.Action+" / "+form.Name)
	c.Success("操作成功")
}
