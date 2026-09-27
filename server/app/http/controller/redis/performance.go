package redis

import (
	"server/app/enum/log_type"
	"server/app/service"
	serviceRedis "server/app/service/software/redis"
	"server/app/util/context"
)

// Performance 获取或更新 Redis 性能配置。
func Performance(c *context.Context) {
	c.SetHeader("Cache-Control", "no-store")
	var form struct {
		Action string                          `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Config *serviceRedis.PerformanceUpdate `json:"config" validate:"omitempty" label:"性能配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Software.Redis().GetPerformance()
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看 Redis 性能配置", "查看 Redis 性能配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		if form.Config == nil {
			c.Error("性能配置不能为空")
			return
		}
		data, err := service.Software.Redis().UpdatePerformance(c.Request.Context(), form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改 Redis 性能配置", "修改 Redis 性能配置")
			c.SuccessWithData("修改成功", data)
		}
	}
}
