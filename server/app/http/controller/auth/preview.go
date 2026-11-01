package auth

import (
	"mime"
	"net/http"
	"server/app/service"
	"server/app/util/context"
	"time"
)

// Preview 按签名输出文件，支持浏览器预览和下载。
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

	data, err := service.File.OpenPreview(form.Key)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer data.File.Close()
	c.SetHeader("Cache-Control", "private, no-store")
	c.SetHeader("Pragma", "no-cache")
	c.SetHeader("Referrer-Policy", "no-referrer")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.SetHeader("Cross-Origin-Resource-Policy", "cross-origin")
	c.SetHeader("Content-Security-Policy", "sandbox")
	contentType := service.File.GetContentType(data.Name)
	c.SetHeader("Content-Type", contentType)
	disposition := "inline"
	if data.Download || !service.File.IsViewableInBrowser(contentType) {
		disposition = "attachment"
	}
	c.SetHeader("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": data.Name}))
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})
	http.ServeContent(c.Writer, c.Request, data.Name, data.ModTime, data.File)
}
