package software

import (
	"errors"
	"net/url"
	"os"
	"server/app/service"
	"server/app/util/context"
	"strings"
)

// Page 读取软件包自带的页面资源，文件访问范围由软件服务限制。
func Page(c *context.Context) {
	var form struct {
		Code string `json:"code" validate:"required,max=100" label:"软件标识"`
		Path string `json:"path" validate:"omitempty,max=1024" label:"资源路径"`
	}
	if ok, msg := c.ValidatorPath(&form); !ok {
		c.Error(msg)
		return
	}
	data, contentType, path, err := service.Software.Page(form.Code, form.Path)
	if err != nil {
		message := err.Error()
		if errors.Is(err, os.ErrNotExist) {
			message = "软件页面资源不存在"
		}
		c.Error(message)
		return
	}
	c.SetHeader("Content-Type", contentType)
	location := url.URL{Path: "/software/" + form.Code + "/page/" + strings.TrimPrefix(path, "web/")}
	c.SetHeader("Content-Location", location.EscapedPath())
	c.SetHeader("Access-Control-Expose-Headers", "Content-Location")
	c.Data(200, data)
}
