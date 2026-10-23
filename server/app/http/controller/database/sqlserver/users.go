package sqlserver

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
		Keyword      string   `json:"keyword" validate:"omitempty,max=1024" label:"账户关键词"`
		Page         int      `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize     int      `json:"page_size" default:"20" validate:"min=1,max=1000" label:"分页容量"`
		Action       string   `json:"action" default:"" validate:"omitempty" label:"操作类型"`
		Database     string   `json:"database" validate:"omitempty,max=1024" label:"授权数据库"`
		Name         string   `json:"name" default:"" validate:"omitempty" label:"账号名称"`
		Password     string   `json:"password" default:"" validate:"omitempty" label:"密码"`
		Privileges   []string `json:"privileges" validate:"omitempty" label:"账号权限"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	// 账户授权范围独立于工作区，未指定时使用当前数据库。
	scope := form.DatabaseName
	if form.Database != "" {
		scope = form.Database
	}
	if form.Action == "password" {
		data, err := service.Database.ChangeSqlServerPassword(ctx, form.Id, scope, form.Name, form.Password)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "SqlServer 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / password")
		c.SuccessWithData("操作成功", data)
		return
	}
	m, err := service.Database.Sqlserver(ctx, form.Id, scope)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	var data interface{}
	switch form.Action {
	case "databases":
		data, err = m.Databases(ctx, form.Keyword, form.Page, form.PageSize, true, "asc")

	case "connection":
		data, err = m.ConnectionInfo(ctx)
	case "", "list":
		data, err = m.Users(ctx, form.Keyword, form.Page, form.PageSize)
	case "create":
		data, err = m.CreateUser(ctx, form.Name, form.Password, form.Privileges)
	case "permissions":
		data, err = m.SetPrivileges(ctx, form.Name, form.Privileges)
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
	if form.Action == "" || form.Action == "list" || form.Action == "databases" || form.Action == "connection" {
		event = log_type.EventShow
	} else if form.Action == "delete" {
		event = log_type.EventDelete
	} else if form.Action == "create" {
		event = log_type.EventCreate
	}
	service.Log.Create(c.GetRequestIp(), event, "SqlServer 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / "+form.Action)
	c.SuccessWithData("操作成功", data)
}
