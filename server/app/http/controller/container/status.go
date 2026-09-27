package container

import (
	"server/app/service"
	"server/app/util/context"
)

// Status 查询容器状态及 CPU、内存、进程、磁盘 I/O 等动态信息。
func Status(c *context.Context) {
	var form struct {
		Id string `json:"id" validate:"required,max=64" label:"容器ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Container.Status(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		c.SuccessWithData("获取成功", data)
	}
}
