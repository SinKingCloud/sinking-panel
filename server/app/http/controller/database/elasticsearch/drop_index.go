package elasticsearch

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// DropIndex 删除索引，使用请求上下文独立限定本次操作。
func (Controller) DropIndex(c *context.Context) {
	var form struct {
		Id           int64    `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string   `json:"database_name" validate:"omitempty,max=255" label:"索引名称"`
		Name         string   `json:"name" validate:"omitempty,max=255" label:"索引名称"`
		Names        []string `json:"names" validate:"omitempty,min=1,dive,required,max=255" label:"索引名称列表"`
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
	var data map[string]interface{}
	if len(form.Names) > 0 {
		err = manager.DropIndices(ctx, form.Names)
	} else {
		data, err = manager.DropIndex(ctx, form.Name)
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "Elasticsearch 删除索引", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	if len(form.Names) > 0 {
		c.Success("删除成功")
	} else {
		c.SuccessWithData("操作成功", data)
	}
}
