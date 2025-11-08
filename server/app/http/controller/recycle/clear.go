package recycle

import (
	"server/app/service"
	"server/app/util/context"
)

// Clear 清空回收站
func Clear(c *context.Context) {
	err := service.Recycle.Clear()
	if err != nil {
		c.Error("清空回收站失败")
	} else {
		c.Success("清空回收站成功")
	}
}
