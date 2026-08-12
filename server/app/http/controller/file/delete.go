package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
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
	f := file.NewDisk("")
	for _, path := range form.Paths {
		if path == "" || !f.Exists(path) {
			c.Error("该目录或文件不存在: " + path)
			return
		}
	}
	for _, path := range form.Paths {
		var err error
		if form.Recycle {
			err = service.Recycle.Create(path)
		} else {
			err = f.Delete(path)
		}
		if err != nil {
			if form.Recycle {
				c.Error("移动至回收站失败")
			} else {
				c.Error("删除失败")
			}
			return
		}
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
