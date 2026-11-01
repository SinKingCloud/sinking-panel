package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取文件信息或分页读取文本内容。
func Info(c *context.Context) {
	var form struct {
		Path     string `json:"path" default:"" validate:"required" label:"文件路径"`
		Read     bool   `json:"read" default:"" validate:"omitempty" label:"是否读取内容"`
		Cursor   int64  `json:"cursor" default:"0" validate:"numeric,min=0" label:"文件游标"`
		PageSize int    `json:"page_size" default:"1000" validate:"numeric,min=1,max=999999" label:"分页容量"`
		Version  string `json:"version" default:"" validate:"omitempty" label:"文件版本"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.File.Info(form.Path, form.Read, form.Cursor, form.PageSize, form.Version)
	if err != nil {
		c.Error(err.Error())
		return
	}
	if form.Cursor == 0 {
		title := "查看文件信息"
		if form.Read {
			title = "读取文件内容"
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, title, title+"["+form.Path+"]")
	}
	c.SuccessWithData("获取成功", data)
}
