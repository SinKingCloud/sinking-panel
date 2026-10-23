package redis

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Value 键值详情。
func (Controller) Value(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Key          string `json:"key" validate:"omitempty,max=65536" label:"键名"`
		Match        string `json:"match" validate:"omitempty,max=65536" label:"匹配条件"`
		Cursor       string `json:"cursor" validate:"omitempty,max=4194304" label:"分页游标"`
		Start        int    `json:"start" validate:"omitempty,min=0" label:"起始位置"`
		PageSize     int    `json:"page_size" validate:"omitempty,min=1,max=500" label:"分页容量"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Redis(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.Value(ctx, form.Key, form.Match, form.Cursor, form.Start, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Redis 键值详情", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
