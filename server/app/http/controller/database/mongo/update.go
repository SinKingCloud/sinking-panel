package mongo

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Update 修改文档。
func (Controller) Update(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Collection   string `json:"collection" validate:"omitempty,max=255" label:"集合"`
		Filter       string `json:"filter" validate:"omitempty,max=1048576" label:"筛选条件"`
		Update       string `json:"update" validate:"omitempty,max=8388608" label:"更新内容"`
		Many         bool   `json:"many"`
		Upsert       bool   `json:"upsert"`
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
	data, err := m.Update(ctx, form.Collection, form.Filter, form.Update, form.Many, form.Upsert)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Mongo 修改文档", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
