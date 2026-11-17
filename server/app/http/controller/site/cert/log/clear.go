package log

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Clear 清理证书续签任务日志。
func Clear(c *context.Context) {
	if err := service.Site.ClearCertLog(); err != nil {
		c.Error("清理日志失败")
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理证书续签日志", "清理证书自动续签任务日志")
	c.Success("清理成功")
}
