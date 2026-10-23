package mysql

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	serviceMysql "server/app/service/database/mysql"
	"server/app/util/context"
	"strconv"
	"strings"
	"time"
)

// Users 账户权限。
func (Controller) Users(c *context.Context) {
	var form struct {
		Id           int64                             `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string                            `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Action       string                            `json:"action" validate:"omitempty" label:"操作类型"`
		Keyword      string                            `json:"keyword" validate:"omitempty,max=1024" label:"账号或来源关键词"`
		Page         int                               `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize     int                               `json:"page_size" default:"20" validate:"min=1,max=1000" label:"每页数量"`
		Database     string                            `json:"database" validate:"omitempty,max=1024" label:"授权数据库"`
		Name         string                            `json:"name" validate:"omitempty" label:"账号名称"`
		NewHost      string                            `json:"new_host" validate:"omitempty,max=255" label:"新的来源地址"`
		Host         string                            `json:"host" validate:"omitempty" label:"来源地址"`
		Password     string                            `json:"password" validate:"omitempty" label:"密码"`
		Grants       []serviceMysql.DatabasePrivileges `json:"grants" validate:"omitempty" label:"数据库权限"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	form.Action = strings.ToLower(strings.TrimSpace(form.Action))
	// 权限配置面向多个数据库，其余账户操作使用打开表单时的数据库。
	scope := form.DatabaseName
	if form.Database != "" && form.Action != "permissions" && form.Action != "update_permissions" {
		scope = form.Database
	}
	if form.Action == "password" {
		data, err := service.Database.ChangeMysqlPassword(ctx, form.Id, scope, form.Name, form.Host, form.Password)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "MySQL 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / password")
		c.SuccessWithData("操作成功", data)
		return
	}
	m, err := service.Database.Mysql(ctx, form.Id, scope)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	var data interface{}
	switch form.Action {
	case "connection":
		data, err = m.ConnectionInfo(ctx)
	case "access":
		data, err = m.ChangeAccess(ctx, form.Name, form.Host, form.NewHost)
	case "", "list":
		data, err = m.Users(ctx, form.Keyword, form.Page, form.PageSize)
	case "create":
		data, err = m.CreateUser(ctx, form.Name, form.Host, form.Password)
	case "permissions":
		data, err = m.Permissions(ctx, form.Name, form.Host)
	case "update_permissions":
		data, err = m.UpdatePermissions(ctx, form.Name, form.Host, form.Grants)
	case "delete":
		data, err = m.DeleteUser(ctx, form.Name, form.Host)
	default:
		c.Error("不支持的账户操作")
		return
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	event := log_type.EventUpdate
	if form.Action == "" || form.Action == "list" || form.Action == "connection" || form.Action == "permissions" {
		event = log_type.EventShow
	} else if form.Action == "delete" {
		event = log_type.EventDelete
	} else if form.Action == "create" {
		event = log_type.EventCreate
	}
	service.Log.Create(c.GetRequestIp(), event, "MySQL 账户权限", "实例["+strconv.FormatInt(form.Id, 10)+"] / "+form.Action)
	c.SuccessWithData("操作成功", data)
}
