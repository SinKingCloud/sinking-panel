package auth

import (
	"server/app/service"
	"server/app/util/context"
)

// Info 网站信息
func Info(c *context.Context) {
	c.SuccessWithData("获取成功", service.System.GetWebInfo())
}
