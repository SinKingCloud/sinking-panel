package elasticsearch

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// CreateIndex 创建索引，使用请求上下文独立限定本次操作。
func (Controller) CreateIndex(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"omitempty,max=255" label:"索引名称"`
		Name         string `json:"name" validate:"omitempty,max=255" label:"索引名称"`
		Settings     string `json:"settings" validate:"omitempty,max=8388608" label:"索引设置"`
		Mapping      string `json:"mapping" validate:"omitempty,max=8388608" label:"字段映射"`
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
	data, err := manager.CreateIndex(ctx, form.Name, form.Settings, form.Mapping)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Elasticsearch 创建索引", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
