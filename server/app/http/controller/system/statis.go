package system

import (
	"server/app/service"
	"server/app/util/context"
)

// Statis 获取系统资源统计。
func Statis(c *context.Context) {
	c.SuccessWithData("获取系统资源统计成功", service.System.GetStatis())
}
