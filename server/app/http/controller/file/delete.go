package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
)

// Delete 删除文件
func Delete(c *context.Context) {
	var form struct {
		Path    string `json:"path" default:"" validate:"required" label:"文件路径"`
		Recycle bool   `json:"recycle" default:"true" validate:"omitempty" label:"软删除"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.Path) {
		c.Error("该目录或文件不存在")
		return
	}
	if !form.Recycle {
		err := f.Delete(form.Path)
		if err != nil {
			c.Error("删除失败")
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除文件", "彻底删除文件["+form.Path+"]")
			c.Success("删除成功")
		}
	} else {
		err := service.Recycle.Create(form.Path)
		if err != nil {
			c.Error("移动至回收站失败")
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除文件", "移动文件至回收站["+form.Path+"]")
			c.Success("移动至回收站成功")
		}
	}
}
