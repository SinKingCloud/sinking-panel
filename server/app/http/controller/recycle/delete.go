package recycle

import (
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除文件
func Delete(c *context.Context) {
	var form struct {
		Name string `json:"name" default:"" validate:"required" label:"文件名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Recycle.Delete(form.Name)
	if err != nil {
		c.Error("彻底删除失败")
	} else {
		c.Success("彻底删除成功")
	}
}
