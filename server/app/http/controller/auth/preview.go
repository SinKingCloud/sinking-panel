package auth

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"server/app/constant"
	"server/app/service"
	"server/app/util/context"
)

// Preview 按签名输出文件，支持浏览器预览和下载。
func Preview(c *context.Context) {
	var form struct {
		Key string `json:"key" default:"" validate:"required,len=32" label:"密钥"`
	}
	// 访问日志读取 RequestURI，校验失败时也不能记录签名。
	c.Request.RequestURI = c.Request.URL.EscapedPath()
	c.SetHeader("Cache-Control", "private, no-store")
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.File.OpenPreview(form.Key)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer data.File.Close()
	contentType := service.File.GetContentType(data.Name)
	disposition := "inline"
	if data.Download || !service.File.IsViewableInBrowser(contentType) {
		disposition = "attachment"
	}
	maxAge := constant.CacheTimeWithFilePreview
	if !data.ExpiresAt.IsZero() {
		maxAge = min(maxAge, time.Until(data.ExpiresAt))
	}
	c.SetHeader("Content-Type", contentType)
	c.SetHeader("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": data.Name}))
	c.SetHeader("Cache-Control", fmt.Sprintf("private, max-age=%d, must-revalidate", max(0, int64(maxAge/time.Second))))
	c.SetHeader("ETag", fmt.Sprintf(`W/"%x-%x"`, data.ModTime.UnixNano(), data.Size))
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})
	http.ServeContent(c.Writer, c.Request, data.Name, data.ModTime, data.File)
}
