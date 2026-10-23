package elasticsearch

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Insert 添加文档，使用请求上下文独立限定本次操作。
func (Controller) Insert(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=255" label:"索引名称"`
		DocumentID   string `json:"document_id" validate:"omitempty,max=512" label:"文档ID"`
		Document     string `json:"document" validate:"required,max=8388608" label:"文档内容"`
		DocumentType string `json:"document_type" validate:"omitempty,max=255" label:"文档类型"`
		Routing      string `json:"routing" validate:"omitempty,max=512" label:"路由值"`
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
	data, err := manager.Insert(ctx, form.DocumentID, form.Document, form.DocumentType, form.Routing)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Elasticsearch 添加文档", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
