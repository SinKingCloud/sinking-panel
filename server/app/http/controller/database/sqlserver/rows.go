package sqlserver

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
		Schema       string `json:"schema" default:"" validate:"omitempty" label:"模式名称"`
		Table        string `json:"table" default:"" validate:"required" label:"数据表"`
		Filters      string `json:"filters" default:"" validate:"omitempty" label:"筛选条件"`
		FilterLogic  string `json:"filter_logic" default:"and" validate:"oneof=and or" label:"筛选关系"`
		OrderBy      string `json:"order_by" default:"" validate:"omitempty" label:"排序字段"`
		Order        string `json:"order" default:"" validate:"omitempty,oneof=asc desc ASC DESC" label:"排序方向"`
		Page         int    `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize     int    `json:"page_size" default:"20" validate:"min=1,max=1000" label:"分页容量"`
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
	data, err := m.Rows(ctx, form.Schema, form.Table, form.Filters, form.FilterLogic, form.OrderBy, form.Order, form.Page, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "SqlServer 分页数据", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
