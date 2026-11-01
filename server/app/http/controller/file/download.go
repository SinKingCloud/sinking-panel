package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Download 创建下载任务。
func Download(c *context.Context) {
	var form struct {
		URL  string `json:"url" default:"" validate:"required,url" label:"远程文件URL"`
		Path string `json:"path" default:"" validate:"required" label:"目标路径"`
		Name string `json:"name" default:"" validate:"omitempty" label:"文件名"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, targetPath, err := service.File.DownloadTask(form.URL, form.Path, form.Name)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建下载任务", "下载文件到["+targetPath+"]")
	c.SuccessWithData("创建下载任务成功", taskID)
}
