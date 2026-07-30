package recycle

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Create 创建文件或目录
func Create(c *context.Context) {
	var form struct {
		Name string `json:"name" default:"" validate:"required" label:"名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Recycle.Create(form.Name)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "移入回收站", "移动文件至回收站["+form.Name+"]")
		c.Success("移动至回收站成功")
	}
}
