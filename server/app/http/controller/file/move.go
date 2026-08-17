package file

import (
	stdContext "context"
	"errors"
	"server/app/enum/log_type"
	"server/app/enum/system_task_status"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"server/app/util/str"
)

func Move(c *context.Context) {
	var form struct {
		SourcePaths []string `json:"source_paths" default:"" validate:"required,min=1,max=1000,unique" label:"源文件路径列表"`
		TargetPath  string   `json:"target_path" default:"" validate:"required" label:"目标路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	for _, sourcePath := range form.SourcePaths {
		if sourcePath == "" || !f.Exists(sourcePath) {
			c.Error("源文件或目录不存在: " + sourcePath)
			return
		}
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	taskName := "移动文件: " + form.SourcePaths[0] + " -> " + form.TargetPath
	taskData := map[string]interface{}{
		"source_paths": append([]string(nil), form.SourcePaths...),
		"target_path":  form.TargetPath,
	}
	service.System.TaskCreate(taskID, taskName, taskData, func(ctx stdContext.Context, rawData interface{}, update func(int, float64, string)) {
		data, ok := rawData.(map[string]interface{})
		if !ok {
			update(system_task_status.Failed, 0, "移动任务数据无效")
			return
		}
		sourcePaths, sourcePathsOK := data["source_paths"].([]string)
		targetPath, targetPathOK := data["target_path"].(string)
		if !sourcePathsOK || !targetPathOK || len(sourcePaths) == 0 || targetPath == "" {
			update(system_task_status.Failed, 0, "移动任务数据无效")
			return
		}
		update(system_task_status.Running, 0, "开始移动")
		var err error
		for index, sourcePath := range sourcePaths {
			err = service.File.MoveWithContext(ctx, sourcePath, targetPath, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
				progress := float64(index) / float64(len(sourcePaths)) * 100
				if total > 0 {
					progress = (float64(index) + float64(current)/float64(total)) / float64(len(sourcePaths)) * 100
				}
				update(system_task_status.Running, progress, "正在移动: "+currentFile)
				return true
			})
			if err != nil {
				break
			}
		}
		if errors.Is(err, stdContext.Canceled) {
			update(system_task_status.Canceled, 0, "任务已取消")
			return
		}
		if err != nil {
			update(system_task_status.Failed, 0, "移动失败: "+err.Error())
			return
		}
		update(system_task_status.Completed, 100, "移动完成")
	})
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建移动任务", taskName)
	c.SuccessWithData("创建移动任务成功", taskID)
}
