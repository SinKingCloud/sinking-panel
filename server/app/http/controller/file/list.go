package file

import (
	"server/app/util/server"
)

func List(c *server.Context) {
	type Form struct {
		Path string `json:"path" default:"/" validate:"required" label:"目录"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	c.SuccessWithData("获取成功", nil)
}
