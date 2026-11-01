package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Sign 获取文件预览或下载签名。
func Sign(c *context.Context) {
	var form struct {
		Path     string `json:"path" default:"" validate:"required" label:"文件路径"`
		Download bool   `json:"download" default:"false" validate:"omitempty" label:"是否下载"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	key, err := service.File.CreatePreviewSign(form.Path, form.Download)
	if err != nil {
		c.Error(err.Error())
	} else {
		title := "预览文件"
		if form.Download {
			title = "下载文件"
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, title, title+"["+form.Path+"]")
		c.SuccessWithData("获取成功", key)
	}
}
