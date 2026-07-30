package recycle

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Clear 清空回收站
func Clear(c *context.Context) {
	err := service.Recycle.Clear()
	if err != nil {
		c.Error("清空回收站失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清空回收站", "清空回收站文件")
		c.Success("清空回收站成功")
	}
}
