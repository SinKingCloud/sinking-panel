package mongo

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Maintenance 维护工具。
func (Controller) Maintenance(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Action       string `json:"action" validate:"omitempty,max=32" label:"操作类型"`
		Collection   string `json:"collection" validate:"omitempty,max=255" label:"集合"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mongo(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	var data interface{}
	switch form.Action {
	case "validate":
		data, err = m.Validate(ctx, form.Collection)
	case "compact":
		data, err = m.Compact(ctx, form.Collection)
	case "drop":
		data, err = m.DropCollection(ctx, form.Collection)
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
	case "check", "validate":
		event = log_type.EventShow
	case "truncate", "drop", "flush":
		event = log_type.EventDelete
	}
	service.Log.Create(c.GetRequestIp(), event, "Mongo 维护工具", "实例["+strconv.FormatInt(form.Id, 10)+"] / "+form.Action)
	c.SuccessWithData("操作成功", data)
}
