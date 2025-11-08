package system

import (
	"server/app/service"
	"server/app/util/context"
)

// Info 获取系统信息
func Info(c *context.Context) {
	// 从服务层获取系统信息
	data := service.System.GetInfo()
	// 返回成功响应
	c.SuccessWithData("获取系统信息成功", data)
}
