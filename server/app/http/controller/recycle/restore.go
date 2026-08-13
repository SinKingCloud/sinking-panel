package recycle

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Restore 恢复文件或目录
func Restore(c *context.Context) {
	var form struct {
		Names []string `json:"names" default:"" validate:"required,min=1,max=1000,unique,dive,required" label:"名称列表"`
		Path  string   `json:"path" default:"" validate:"omitempty" label:"新目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	for _, name := range form.Names {
		_ = service.Recycle.Restore(name, form.Path)
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "恢复回收站文件", "恢复回收站文件")
	c.Success("恢复成功")
}
