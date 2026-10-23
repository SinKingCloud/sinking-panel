package sqlserver

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Objects 数据库对象。
func (Controller) Objects(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Keyword      string `json:"keyword" validate:"omitempty,max=1024" label:"名称关键词"`
		Page         int    `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize     int    `json:"page_size" default:"20" validate:"min=1,max=1000" label:"分页容量"`
		Kind         string `json:"kind" default:"view" validate:"required,oneof=view routine trigger foreign_key" label:"对象类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Sqlserver(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.SelectObjects(ctx, form.Schema, form.Kind, form.Keyword, form.Page, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "SqlServer 数据库对象", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
