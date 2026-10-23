package elasticsearch

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Documents 搜索文档，使用请求上下文独立限定本次操作。
func (Controller) Documents(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=255" label:"索引名称"`
		Query        string `json:"query" validate:"omitempty,max=8388608" label:"搜索DSL"`
		Pagination   string `json:"pagination" validate:"omitempty,oneof=offset cursor pit scroll" label:"分页方式"`
		SearchID     string `json:"search_id" validate:"omitempty,max=131072" label:"分页会话"`
		Cursor       string `json:"cursor" validate:"omitempty,max=65536" label:"分页游标"`
		Page         int    `json:"page" default:"1" validate:"omitempty,min=1,max=1000000000" label:"页码"`
		PageSize     int    `json:"page_size" default:"50" validate:"omitempty,min=1,max=500" label:"分页容量"`
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
	data, err := manager.Documents(ctx, form.Query, form.Pagination, form.SearchID, form.Cursor, form.Page, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Elasticsearch 搜索文档", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
