package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Copy 创建复制任务。
func Copy(c *context.Context) {
	var form struct {
		SourcePaths []string `json:"source_paths" default:"" validate:"required,min=1,max=1000,unique" label:"源文件路径列表"`
		TargetPath  string   `json:"target_path" default:"" validate:"required" label:"目标路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.File.CopyTask(form.SourcePaths, form.TargetPath)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建复制任务", "复制文件: "+form.SourcePaths[0]+" -> "+form.TargetPath)
	c.SuccessWithData("创建复制任务成功", taskID)
}
