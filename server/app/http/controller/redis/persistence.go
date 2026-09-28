package redis

import (
	"server/app/enum/log_type"
	"server/app/service"
	serviceRedis "server/app/service/software/redis"
	"server/app/util/context"
)

// Persistence 获取或更新 Redis 持久化配置。
func Persistence(c *context.Context) {
	var form struct {
		Action string                          `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Config *serviceRedis.PersistenceUpdate `json:"config" validate:"omitempty" label:"持久化配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Software.Redis().GetPersistence()
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看 Redis 持久化配置", "查看 Redis 持久化配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		data, err := service.Software.Redis().UpdatePersistence(c.Request.Context(), form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改 Redis 持久化配置", "修改 Redis 持久化配置")
			c.SuccessWithData("修改成功", data)
		}
	}
}
