package recycle

import (
	"server/app/service"
	"server/app/util/server"
)

// Create 创建文件或目录
func Create(c *server.Context) {
	type Form struct {
		Name string `json:"name" default:"" validate:"required" label:"名称"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	err := service.Recycle.Create(form.Name)
	if err != nil {
		c.Error(err.Error())
	} else {
		c.Success("移动至回收站成功")
	}
}
