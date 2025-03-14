package file

import (
	"server/app/util/file"
	"server/app/util/server"
	"time"
)

func Info(c *server.Context) {
	type Form struct {
		Path     string `json:"path" default:"" validate:"required" label:"文件路径"`
		Read     bool   `json:"read" default:"" validate:"omitempty" label:"是否读取内容"`
		Page     int    `json:"page" default:"" validate:"omitempty,numeric,min=1" label:"分页页码"`
		PageSize int    `json:"page_size" default:"" validate:"omitempty,numeric,max=999999" label:"分页容量"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
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
		data["content"], _ = f.GetFileContent(form.Path, form.Page, form.PageSize)
	}
	c.SuccessWithData("获取成功", data)
}
