package file

import (
	"server/app/util/context"
	"server/app/util/file"
	"time"
)

func Info(c *context.Context) {
	var form struct {
		Path     string `json:"path" default:"" validate:"required" label:"文件路径"`
		Read     bool   `json:"read" default:"" validate:"omitempty" label:"是否读取内容"`
		Page     int    `json:"page" default:"1" validate:"numeric,min=1" label:"分页页码"`
		PageSize int    `json:"page_size" default:"1000" validate:"numeric,min=1,max=999999" label:"分页容量"`
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
	info, err := f.FileInfo(form.Path)
	if err != nil {
		c.Error("获取失败")
		return
	}
	data := map[string]interface{}{
		"name":        info.Name,
		"size":        info.Size,
		"mode":        info.Mode,
		"is_dir":      info.IsDir,
		"update_time": time.Unix(info.UpdateTime, 0).Format("2006-01-02 15:04:05"),
	}
	if form.Read {
		content, err := f.GetFileContent(form.Path, form.Page, form.PageSize)
		if err != nil {
			c.Error("读取文件内容失败")
			return
		}
		data["content"] = content
	}
	c.SuccessWithData("获取成功", data)
}
