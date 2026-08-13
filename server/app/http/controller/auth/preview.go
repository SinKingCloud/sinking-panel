package auth

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"server/app/service"
	"server/app/util/context"
	"time"
)

func Preview(c *context.Context) {
	var form struct {
		Key string `json:"key" default:"" validate:"required,len=32" label:"密钥"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	// 避免临时密钥写入应用访问日志
	c.Request.URL.RawQuery = ""
	c.Request.RequestURI = c.Request.URL.RequestURI()

	sign, err := service.File.CheckPreviewSign(form.Key)
	if err != nil {
		c.Error("预览链接无效或已过期")
		return
	}
	info, err := os.Stat(sign.Path)
	if err != nil {
		c.Error("文件不存在")
		return
	}
	if info.IsDir() {
		c.Error("不允许访问目录")
		return
	}
	if !info.Mode().IsRegular() {
		c.Error("不支持访问该文件类型")
		return
	}
	fileHandle, err := os.Open(sign.Path)
	if err != nil {
		c.Error("无法访问文件")
		return
	}
	defer fileHandle.Close()
	info, err = fileHandle.Stat()
	if err != nil {
		c.Error("无法访问文件")
		return
	}
	if info.IsDir() {
		c.Error("不允许访问目录")
		return
	}
	if !info.Mode().IsRegular() {
		c.Error("不支持访问该文件类型")
		return
	}
	fileName := filepath.Base(sign.FileName)
	if fileName == "." || fileName == string(filepath.Separator) {
		fileName = filepath.Base(sign.Path)
	}
	c.SetHeader("Cache-Control", "private, no-store")
	c.SetHeader("Pragma", "no-cache")
	c.SetHeader("Referrer-Policy", "no-referrer")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.SetHeader("Cross-Origin-Resource-Policy", "cross-origin")
	c.SetHeader("Content-Security-Policy", "sandbox")
	contentType := service.File.GetContentType(fileName)
	c.SetHeader("Content-Type", contentType)
	disposition := "inline"
	if sign.Download || !service.File.IsViewableInBrowser(contentType) {
		disposition = "attachment"
	}
	c.SetHeader("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": fileName}))
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})
	http.ServeContent(c.Writer, c.Request, fileName, info.ModTime(), fileHandle)
}
