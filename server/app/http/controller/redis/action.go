package redis

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Uninstall 停止并删除 Redis 容器及其专属数据。
func Uninstall(c *context.Context) {
	if err := service.Software.Redis().Uninstall(c.Request.Context()); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "卸载 Redis", "停止并删除 Redis 容器及其配置数据")
		c.Success("卸载成功")
	}
}

// Start 启动 Redis 容器。
func Start(c *context.Context) {
	if _, err := service.Container.Start(c.Request.Context(), service.Software.Redis().InstanceID); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "启动 Redis", "启动 Redis 容器")
		c.Success("启动成功")
	}
}

// Stop 停止 Redis 容器。
func Stop(c *context.Context) {
	if err := service.Container.Stop(c.Request.Context(), service.Software.Redis().InstanceID); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "停止 Redis", "停止 Redis 容器")
		c.Success("停止成功")
	}
}

// Restart 重启 Redis 容器。
func Restart(c *context.Context) {
	if err := service.Container.Restart(c.Request.Context(), service.Software.Redis().InstanceID); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "重启 Redis", "重启 Redis 容器")
		c.Success("重启成功")
	}
}

// Reload 重新加载 Redis 配置。
func Reload(c *context.Context) {
	data, err := service.Software.Redis().Reload(c.Request.Context())
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "重载 Redis", "重新加载 Redis 配置")
		c.SuccessWithData("重载成功", data)
	}
}
