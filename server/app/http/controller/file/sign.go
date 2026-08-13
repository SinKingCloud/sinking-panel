package file

import (
	"os"
	"path/filepath"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
)

func Sign(c *context.Context) {
	var form struct {
		Path     string `json:"path" default:"" validate:"required" label:"文件路径"`
		Download bool   `json:"download" default:"false" validate:"omitempty" label:"是否下载"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.Path) {
		c.Error("文件不存在")
		return
	}
	if f.IsDir(form.Path) {
		c.Error("不能预览目录")
		return
	}
	filePath := f.Path(form.Path, true, true)
	fileName := filepath.Base(filePath)
	filePath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		c.Error("文件不存在")
		return
	}
	info, err := os.Stat(filePath)
	if err != nil {
		c.Error("文件不存在")
		return
	}
	if !info.Mode().IsRegular() {
		c.Error("不支持访问该文件类型")
		return
	}
	key, err := service.File.CreatePreviewSign(filePath, fileName, form.Download)
	if err != nil {
		c.Error(err.Error())
		return
	}
	title := "预览文件"
	if form.Download {
		title = "下载文件"
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, title, title+"["+form.Path+"]")
	c.SuccessWithData("获取成功", key)
}
