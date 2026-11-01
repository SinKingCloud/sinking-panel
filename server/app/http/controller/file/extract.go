package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Extract 创建解压任务。
func Extract(c *context.Context) {
	var form struct {
		Path string `json:"path" default:"" validate:"required" label:"压缩文件路径"`
		Dir  string `json:"dir" default:"" validate:"required" label:"解压目标目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.File.ExtractTask(form.Path, form.Dir)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建解压任务", "解压文件: "+form.Path)
	c.SuccessWithData("创建解压任务成功", taskID)
}
