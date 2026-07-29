package file

import (
	"server/app/util/context"
	"server/app/util/file"
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
	f := file.NewDisk(form.Path)
	if f.Exists(form.Name) {
		c.Error("该目录或文件已存在")
		return
	}
	err := f.AutoCreate(form.Name)
	if err != nil {
		c.Error("创建目录或文件失败")
		return
	}
	if form.Chmod > 0 {
		_ = f.SetFileMode(form.Name, form.Chmod)
	}
	c.Success("创建成功")
}
