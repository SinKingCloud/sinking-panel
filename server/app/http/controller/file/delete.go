package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Delete 删除文件
func Delete(c *context.Context) {
	var form struct {
		Paths   []string `json:"paths" default:"" validate:"required,min=1,max=1000,unique" label:"文件路径列表"`
		Recycle bool     `json:"recycle" default:"true" validate:"omitempty" label:"软删除"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.File.Delete(form.Paths, form.Recycle); err != nil {
		c.Error(err.Error())
		return
	}
	detail := form.Paths[0]
	if len(form.Paths) > 1 {
		detail = strconv.Itoa(len(form.Paths)) + "项"
	}
	if form.Recycle {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除文件", "移动文件至回收站["+detail+"]")
		c.Success("移动至回收站成功")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除文件", "彻底删除文件["+detail+"]")
		c.Success("删除成功")
	}
}
