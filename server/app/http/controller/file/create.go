package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Create 创建文件或目录
func Create(c *context.Context) {
	var form struct {
		Name  string `json:"name" default:"" validate:"required" label:"名称"`
		Path  string `json:"path" default:"" validate:"required" label:"目录"`
		Chmod uint32 `json:"chmod" default:"0" validate:"omitempty,numeric" label:"权限"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.File.Create(form.Path, form.Name, form.Chmod)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建文件", "创建文件或目录["+form.Path+"/"+form.Name+"]")
		c.Success("创建成功")
	}
}
