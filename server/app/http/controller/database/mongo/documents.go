package mongo

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Documents 文档列表。
func (Controller) Documents(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Collection   string `json:"collection" validate:"omitempty,max=255" label:"集合"`
		Filter       string `json:"filter" validate:"omitempty,max=1048576" label:"筛选条件"`
		Sort         string `json:"sort" validate:"omitempty,max=65536" label:"排序条件"`
		Projection   string `json:"projection" validate:"omitempty,max=65536" label:"显示字段"`
		Page         int    `json:"page" default:"1" validate:"omitempty,min=1,max=1000000" label:"页码"`
		PageSize     int    `json:"page_size" default:"50" validate:"omitempty,min=1,max=1000" label:"分页容量"`
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
	data, err := m.Documents(ctx, form.Collection, form.Filter, form.Sort, form.Projection, form.Page, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Mongo 文档列表", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
