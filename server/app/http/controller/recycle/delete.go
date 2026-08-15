package recycle

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除文件
func Delete(c *context.Context) {
	var form struct {
		Names []string `json:"names" default:"" validate:"required,min=1,max=1000,unique,dive,required" label:"文件名称列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	for _, name := range form.Names {
		if err := service.Recycle.Delete(name); err != nil {
			c.Error("彻底删除失败: " + err.Error())
			return
		}
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "彻底删除文件", "彻底删除回收站文件")
	c.Success("彻底删除成功")
}
