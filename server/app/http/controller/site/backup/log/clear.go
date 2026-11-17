package log

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Clear 清理网站备份任务日志。
func Clear(c *context.Context) {
	if err := service.Site.ClearBackupLog(); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理网站备份日志", "清理网站自动备份任务日志")
	c.Success("清理成功")
}
