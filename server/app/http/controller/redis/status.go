package redis

import (
	"server/app/service"
	"server/app/util/context"
)

// Status 获取 Redis 容器及服务状态。
func Status(c *context.Context) {
	data, err := service.Software.Redis().Status(c.Request.Context())
	if err != nil {
		c.Error(err.Error())
	} else {
		c.SuccessWithData("获取成功", data)
	}
}
