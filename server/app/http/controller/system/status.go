package system

import (
	"server/app/service"
	"server/app/util/context"
)

// Status 获取系统状态信息（网卡流量和磁盘IO）
func Status(c *context.Context) {
	var form struct {
		Interface string `json:"interface" default:"" validate:"omitempty" label:"网卡名称"`
		Disk      string `json:"disk" default:"" validate:"omitempty" label:"磁盘名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := service.System.GetStatus(form.Interface, form.Disk)
	c.SuccessWithData("获取系统状态成功", data)
}
