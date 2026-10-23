package postgres

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	servicePostgres "server/app/service/database/postgres"
	"server/app/util/context"
	"strconv"
	"strings"
	"time"
)

// Users 账户权限。
func (Controller) Users(c *context.Context) {
	var form struct {
		Id           int64                                `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string                               `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Keyword      string                               `json:"keyword" validate:"omitempty,max=1024" label:"账户关键词"`
		Page         int                                  `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize     int                                  `json:"page_size" default:"20" validate:"min=1,max=1000" label:"分页容量"`
		Action       string                               `json:"action" default:"" validate:"omitempty" label:"操作类型"`
		Database     string                               `json:"database" validate:"omitempty,max=1024" label:"授权数据库"`
		Name         string                               `json:"name" default:"" validate:"omitempty" label:"账号名称"`
		Password     string                               `json:"password" default:"" validate:"omitempty" label:"密码"`
		Privileges   []string                             `json:"privileges" validate:"omitempty" label:"账号权限"`
		Grants       []servicePostgres.DatabasePrivileges `json:"grants" validate:"omitempty" label:"数据库权限"`
		Schema       string                               `json:"schema" default:"" validate:"omitempty" label:"模式名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	form.Action = strings.ToLower(strings.TrimSpace(form.Action))
	// 账户授权范围独立于工作区，未指定时使用当前数据库。
	scope := form.DatabaseName
	if form.Database != "" && form.Action != "permissions_info" && form.Action != "update_permissions" {
		scope = form.Database
	}
	if form.Action == "password" {
		data, err := service.Database.ChangePostgresPassword(ctx, form.Id, scope, form.Name, form.Password)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Postgres 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / password")
		c.SuccessWithData("操作成功", data)
		return
	}
	m, err := service.Database.Postgres(ctx, form.Id, scope)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	var data interface{}
	switch form.Action {
	case "databases":
		data, err = m.Databases(ctx, form.Keyword, form.Page, form.PageSize, false, "asc")

	case "connection":
		data, err = m.ConnectionInfo(ctx)
	case "schemas":
		data, err = m.Schemas(ctx)
	case "", "list":
		data, err = m.Users(ctx, form.Schema, form.Keyword, form.Page, form.PageSize)
	case "create":
		data, err = m.CreateUser(ctx, form.Name, form.Password)
	case "permissions_info":
		data, err = m.Permissions(ctx, form.Name)
	case "update_permissions":
		data, err = m.UpdatePermissions(ctx, form.Name, form.Grants)
	case "grant":
		data, err = m.Grant(ctx, form.Name, form.Schema, form.Privileges)
	case "revoke":
		data, err = m.Revoke(ctx, form.Name, form.Schema, form.Privileges)
	case "delete":
		data, err = m.DeleteUser(ctx, form.Name, form.Schema)
	default:
		c.Error("不支持的账户操作")
		return
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	event := log_type.EventUpdate
	if form.Action == "" || form.Action == "list" || form.Action == "databases" || form.Action == "connection" || form.Action == "schemas" || form.Action == "permissions_info" {
		event = log_type.EventShow
	} else if form.Action == "delete" {
		event = log_type.EventDelete
	} else if form.Action == "create" {
		event = log_type.EventCreate
	}
	service.Log.Create(c.GetRequestIp(), event, "Postgres 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / "+form.Action)
	c.SuccessWithData("操作成功", data)
}
