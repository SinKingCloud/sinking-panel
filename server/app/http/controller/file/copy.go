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
	"time"
)

func Copy(c *context.Context) {
	var form struct {
		SourcePath string `json:"source_path" default:"" validate:"required" label:"源文件路径"`
		TargetPath string `json:"target_path" default:"" validate:"required" label:"目标路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.SourcePath) {
		c.Error("源文件或目录不存在")
		return
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	taskName := "复制文件: " + form.SourcePath + " -> " + form.TargetPath
	taskData := map[string]interface{}{
		"source_path": form.SourcePath,
		"target_path": form.TargetPath,
	}
	_ = service.System.TaskCreate(taskID, taskName, taskData)
	service.System.SetTaskCancelFunc(taskID, nil)
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			service.System.TaskDelete(taskID)
		}()
		taskInfo := service.System.GetTask(taskID)
		if taskInfo == nil || taskInfo.Context == nil {
			return
		}
		service.System.TaskUpdate(taskID, system_task_status.Running, 0, "开始复制")
		err := service.File.CopyWithContext(taskInfo.Context, form.SourcePath, form.TargetPath, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			message := "正在复制: " + currentFile
			service.System.TaskUpdate(taskID, system_task_status.Running, progress, message)
			return true
		})
		if errors.Is(err, stdContext.Canceled) {
			service.System.TaskUpdate(taskID, system_task_status.Canceled, 0, "任务已取消")
			return
		}
		if err != nil {
			service.System.TaskUpdate(taskID, system_task_status.Failed, 0, "复制失败: "+err.Error())
			return
		}
		service.System.TaskUpdate(taskID, system_task_status.Completed, 100, "复制完成")
	}()
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建复制任务", taskName)
	c.SuccessWithData("创建复制任务成功", taskID)
}
