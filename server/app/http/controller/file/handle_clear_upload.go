package file

import (
	"server/app/service"
	"server/app/util/context"
)

func handleClearUpload(c *context.Context) {
	err := service.File.ClearUpload(c.DefaultForm("upload_id", ""))
	if err != nil {
		c.Error(err.Error())
	} else {
		c.Success("清理上传缓存成功")
	}
}
