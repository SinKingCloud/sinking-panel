package recycle

import (
	"server/app/service"
	"server/app/util/context"
)

// Restore 恢复文件或目录
func Restore(c *context.Context) {
	var form struct {
		Name string `json:"name" default:"" validate:"required" label:"名称"`
		Path string `json:"path" default:"" validate:"omitempty" label:"新目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Recycle.Restore(form.Name, form.Path)
	if err != nil {
		c.Error(err.Error())
	} else {
		c.Success("恢复成功")
	}
}
