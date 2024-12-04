package auth

import (
	"server/app/util/server"
)

type ControllerWeb struct {
}

func (ControllerWeb) Info(c *server.Context) {
	c.Success("请求成功")
}

func (ControllerWeb) Login(c *server.Context) {
	c.Success("success")
}

// OutLogin 退出登录
func (ControllerWeb) OutLogin(c *server.Context) {
	c.Success("success")
}
