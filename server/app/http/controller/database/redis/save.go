package redis

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Save 保存键值。
func (Controller) Save(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Key          string `json:"key" validate:"omitempty,max=65536" label:"键名"`
		Type         string `json:"type" validate:"omitempty,max=32" label:"数据类型"`
		Value        string `json:"value" validate:"omitempty,max=8388608" label:"键值"`
		TTL          *int64 `json:"ttl" validate:"omitempty,min=-1,max=9223372036" label:"有效期"`
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
	data, err := m.Save(ctx, form.Key, form.Type, form.Value, form.TTL)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Redis 保存键值", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
