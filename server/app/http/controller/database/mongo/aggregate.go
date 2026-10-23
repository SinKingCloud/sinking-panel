package mongo

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Aggregate 聚合查询。
func (Controller) Aggregate(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Collection   string `json:"collection" validate:"omitempty,max=255" label:"集合"`
		Pipeline     string `json:"pipeline" validate:"omitempty,max=1048576" label:"聚合管道"`
		Limit        int    `json:"limit" validate:"omitempty,min=1,max=1000" label:"返回上限"`
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
	data, err := m.Aggregate(ctx, form.Collection, form.Pipeline, form.Limit)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Mongo 聚合查询", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
