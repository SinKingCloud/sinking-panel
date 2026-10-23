package redis

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Users 账户权限。
func (Controller) Users(c *context.Context) {
	var form struct {
		Id           int64    `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string   `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Action       string   `json:"action" validate:"omitempty,max=32" label:"操作类型"`
		Name         string   `json:"name" validate:"omitempty,max=128" label:"账户名称"`
		Password     string   `json:"password" validate:"omitempty,max=1024" label:"密码"`
		Privileges   []string `json:"privileges" validate:"omitempty,max=16,dive,max=64" label:"权限"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Redis(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	var data interface{}
	switch form.Action {
	case "connection":
		data, err = m.ConnectionInfo(ctx)
	case "", "list":
		data, err = m.Users(ctx)
	case "create":
		data, err = m.CreateUser(ctx, form.Name, form.Password, form.Privileges)
	case "permissions":
		data, err = m.Permissions(ctx, form.Name, form.Privileges)
	case "password":
		data, err = m.ChangePassword(ctx, form.Name, form.Password)
	case "grant":
		data, err = m.Grant(ctx, form.Name, form.Privileges)
	case "revoke":
		data, err = m.Revoke(ctx, form.Name, form.Privileges)
	case "delete":
		data, err = m.DeleteUser(ctx, form.Name)
	default:
		c.Error("不支持的账户操作")
		return
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	event := log_type.EventUpdate
	if form.Action == "" || form.Action == "list" || form.Action == "connection" {
		event = log_type.EventShow
	} else if form.Action == "delete" {
		event = log_type.EventDelete
	} else if form.Action == "create" {
		event = log_type.EventCreate
	}
	service.Log.Create(c.GetRequestIp(), event, "Redis 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / "+form.Action)
	c.SuccessWithData("操作成功", data)
}
