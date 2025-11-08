package file

import (
	"server/app/util/context"
	"server/app/util/file"
)

// Count 统计信息
func Count(c *context.Context) {
	type Form struct {
		Path string `json:"path" default:"" validate:"required" label:"文件路径"`
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
	totalSize, fileCount, dirCount, err := f.Count(form.Path)
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", map[string]int64{
		"size": totalSize,
		"file": fileCount,
		"dir":  dirCount,
	})
}
