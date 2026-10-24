package types

import (
	"server/app/service"
	"server/app/util/context"
)

// Enum 获取指定模块的分类枚举。
func Enum(c *context.Context) {
	var form struct {
		Module string `json:"module" default:"" validate:"required,max=50" label:"所属模块"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Type.GetEnum(form.Module)
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
