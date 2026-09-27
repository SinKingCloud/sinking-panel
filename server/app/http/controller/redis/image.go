package redis

import (
	"path/filepath"
	"strings"

	"server/app/service"
	"server/app/util/context"
)

// Image 导入本机镜像归档或查询已导入镜像，返回 Redis 安装预览及默认配置。
func Image(c *context.Context) {
	var form struct {
		ImagePath string `json:"image_path" default:"" validate:"required" label:"镜像路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	source := strings.TrimSpace(form.ImagePath)
	if filepath.IsAbs(source) {
		image, err := service.Container.Import(source)
		if err != nil {
			c.Error(err.Error())
			return
		}
		source = image.Id
	}
	data, err := service.Container.Redis().Image(source)
	if err != nil {
		c.Error(err.Error())
	} else {
		c.SuccessWithData("解析成功", data)
	}
}
