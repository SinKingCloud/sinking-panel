package file

import (
	"server/app/util/file"
	"server/app/util/server"
)

func List(c *server.Context) {
	type Form struct {
		Path string `json:"path" default:"/" validate:"required" label:"目录"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	list, err := f.FileList(form.Path)
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", list)
}
