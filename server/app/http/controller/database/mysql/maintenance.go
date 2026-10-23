package mysql

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
	"time"
)

// Maintenance 维护工具。
func (Controller) Maintenance(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Action       string `json:"action" validate:"omitempty" label:"操作类型"`
		Schema       string `json:"schema" validate:"omitempty" label:"数据库架构"`
		Table        string `json:"table" validate:"omitempty" label:"数据表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mysql(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	var data interface{}
	form.Action = strings.ToLower(strings.TrimSpace(form.Action))
	switch form.Action {
	case "check":
		data, err = m.Check(ctx, form.Schema, form.Table)
	case "analyze":
		data, err = m.Analyze(ctx, form.Schema, form.Table)
	case "optimize":
		data, err = m.Optimize(ctx, form.Schema, form.Table)
	case "repair":
		data, err = m.Repair(ctx, form.Schema, form.Table)
	case "truncate":
		data, err = m.Truncate(ctx, form.Schema, form.Table)
	case "drop":
		data, err = m.Drop(ctx, form.Schema, form.Table)
	default:
		c.Error("不支持的维护操作")
		return
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	event := log_type.EventUpdate
	switch form.Action {
	case "check":
		event = log_type.EventShow
	case "truncate", "drop":
		event = log_type.EventDelete
	}
	service.Log.Create(c.GetRequestIp(), event, "MySQL 维护工具", "实例["+strconv.FormatInt(form.Id, 10)+"] / "+form.Action)
	c.SuccessWithData("操作成功", data)
}
