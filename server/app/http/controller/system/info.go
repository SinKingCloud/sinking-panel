package system

import (
	"server/app/service"
	"server/app/util/context"
)

// Info 获取主机、操作系统和运行时的基础信息。
func Info(c *context.Context) {
	c.SuccessWithData("获取数据成功", service.System.GetInfo())
}
