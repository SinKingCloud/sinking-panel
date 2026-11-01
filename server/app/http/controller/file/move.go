package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Move 创建移动任务。
func Move(c *context.Context) {
	var form struct {
		SourcePaths []string `json:"source_paths" default:"" validate:"required,min=1,max=1000,unique" label:"源文件路径列表"`
		TargetPath  string   `json:"target_path" default:"" validate:"required" label:"目标路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.File.MoveTask(form.SourcePaths, form.TargetPath)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建移动任务", "移动文件: "+form.SourcePaths[0]+" -> "+form.TargetPath)
	c.SuccessWithData("创建移动任务成功", taskID)
}
