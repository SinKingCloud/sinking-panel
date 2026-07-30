package recycle

import (
	"server/app/enum/log_type"
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
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "彻底删除文件", "彻底删除回收站文件")
		c.Success("彻底删除成功")
	}
}
