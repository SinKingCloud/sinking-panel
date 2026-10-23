package elasticsearch

import (
	stdContext "context"
	"server/app/service"
	"server/app/util/context"
	"time"
)

// CloseSearch 释放文档浏览创建的分页会话。
func (Controller) CloseSearch(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=255" label:"索引名称"`
		Pagination   string `json:"pagination" validate:"required,oneof=pit scroll" label:"分页方式"`
		SearchID     string `json:"search_id" validate:"required,max=131072" label:"分页会话"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	manager, err := service.Database.Elasticsearch(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer manager.Close()
	if err := manager.CloseSearch(ctx, form.Pagination, form.SearchID); err != nil {
		c.Error(err.Error())
		return
	}
	c.Success("操作成功")
}
