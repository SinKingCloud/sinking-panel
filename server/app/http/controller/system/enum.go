package system

import (
	"server/app/enum"
	"server/app/util/context"
)

func Enum(c *context.Context) {
	var form struct {
		Name string `json:"name" default:"" validate:"required,name" label:"枚举名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if value, ok := enum.Data[form.Name]; ok {
		c.SuccessWithData("获取数据成功", value)
	} else {
		c.Error("枚举类型不存在")
	}
}
