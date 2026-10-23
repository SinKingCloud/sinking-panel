package mysql

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Rows 分页数据。
func (Controller) Rows(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty" label:"数据库架构"`
		Table        string `json:"table" validate:"omitempty" label:"数据表"`
		Filters      string `json:"filters" validate:"omitempty" label:"筛选条件"`
		FilterLogic  string `json:"filter_logic" default:"and" validate:"omitempty,oneof=and or" label:"筛选方式"`
		OrderBy      string `json:"order_by" validate:"omitempty" label:"排序字段"`
		Order        string `json:"order" validate:"omitempty,oneof=asc desc ASC DESC" label:"排序方向"`
		Page         int    `json:"page" default:"1" validate:"omitempty,min=0,max=1000000" label:"页码"`
		PageSize     int    `json:"page_size" default:"20" validate:"omitempty,min=0" label:"每页数量"`
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
	data, err := m.Rows(ctx, form.Schema, form.Table, form.Filters, form.FilterLogic, form.OrderBy, form.Order, form.Page, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "MySQL 分页数据", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
