package system

import (
	"server/app/service"
	"server/app/util/context"
)

// IO 获取网卡和磁盘的 I/O 统计及采样历史。
func IO(c *context.Context) {
	var form struct {
		Interface string `json:"interface" default:"" validate:"omitempty" label:"网卡名称"`
		Disk      string `json:"disk" default:"" validate:"omitempty" label:"磁盘名称"`
		After     int64  `json:"after" default:"0" validate:"omitempty,min=0" label:"采样游标"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	c.SuccessWithData("获取系统 I/O 信息成功", service.System.GetIO(form.Interface, form.Disk, form.After))
}
