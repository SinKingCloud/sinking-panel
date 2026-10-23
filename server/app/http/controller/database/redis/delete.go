package redis

import (
	stdContext "context"
	"encoding/json"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Delete 删除键值。
func (Controller) Delete(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Keys         string `json:"keys" validate:"omitempty,max=1048576" label:"键名列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var keys []string
	if json.Unmarshal([]byte(form.Keys), &keys) != nil || len(keys) == 0 || len(keys) > 1000 {
		c.Error("请选择 1 至 1000 个键")
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
	data, err := m.Delete(ctx, keys)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "Redis 删除键值", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
