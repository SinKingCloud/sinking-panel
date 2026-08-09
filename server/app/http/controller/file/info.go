package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"time"
)

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
		content, nextCursor, eof, version, err := f.GetFileContent(form.Path, form.Cursor, form.PageSize, form.Version)
		if err != nil {
			c.Error("读取文件内容失败: " + err.Error())
			return
		}
		data["content"] = content
		data["cursor"] = form.Cursor
		data["next_cursor"] = nextCursor
		data["eof"] = eof
		data["version"] = version
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
