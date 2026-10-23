package elasticsearch

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Indices 索引列表，使用请求上下文独立限定本次操作。
func (Controller) Indices(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"omitempty,max=255" label:"索引名称"`
		Keyword      string `json:"keyword" validate:"omitempty,max=255" label:"索引筛选"`
		OrderByType  string `json:"order_by_type" default:"asc" validate:"omitempty,oneof=asc desc" label:"排序类型"`
		Page         int    `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize     int    `json:"page_size" default:"20" validate:"min=1,max=1000" label:"每页数量"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	manager, err := service.Database.Elasticsearch(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer manager.Close()
	data, err := manager.Indices(ctx, form.Keyword, form.Page, form.PageSize, form.OrderByType)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Elasticsearch 索引列表", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
