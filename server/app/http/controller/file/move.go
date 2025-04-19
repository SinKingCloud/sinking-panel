package file

import (
	"server/app/service"
	"server/app/service/system"
	"server/app/util/file"
	"server/app/util/server"
	"server/app/util/str"
	"sync/atomic"
	"time"
)

func Move(c *server.Context) {
	type Form struct {
		SourcePath string `json:"source_path" default:"" validate:"required" label:"源文件路径"`
		TargetPath string `json:"target_path" default:"" validate:"required" label:"目标路径"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.SourcePath) {
		c.Error("源文件或目录不存在")
		return
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	taskName := "移动文件: " + form.SourcePath + " -> " + form.TargetPath
	taskData := map[string]interface{}{
		"source_path": form.SourcePath,
		"target_path": form.TargetPath,
	}
	_ = service.System.TaskCreate(taskID, taskName, taskData)
	var canceled atomic.Bool
	service.System.SetTaskCancelFunc(taskID, func() {
		canceled.Store(true)
	})
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			service.System.TaskDelete(taskID)
		}()
		taskInfo := service.System.GetTask(taskID)
		if taskInfo == nil || taskInfo.Context == nil {
			return
		}
		service.System.TaskUpdate(taskID, system.TaskStatusRunning, 0, "开始移动")
		_, _, _, err := f.Count(form.SourcePath)
		if err != nil {
			service.System.TaskUpdate(taskID, system.TaskStatusFailed, 0, "计算文件大小失败: "+err.Error())
			return
		}
		err = service.File.MoveWithContext(taskInfo.Context, form.SourcePath, form.TargetPath, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if canceled.Load() {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			message := "正在移动: " + currentFile
			service.System.TaskUpdate(taskID, system.TaskStatusRunning, progress, message)
			return true
		})
		select {
		case <-taskInfo.Context.Done():
			service.System.TaskUpdate(taskID, system.TaskStatusCanceled, 0, "任务已取消")
			return
		default:
			if canceled.Load() {
				service.System.TaskUpdate(taskID, system.TaskStatusCanceled, 0, "任务已取消")
				return
			}
			if err != nil {
				service.System.TaskUpdate(taskID, system.TaskStatusFailed, 0, "移动失败: "+err.Error())
				return
			}
			service.System.TaskUpdate(taskID, system.TaskStatusCompleted, 100, "移动完成")
		}
	}()
	c.SuccessWithData("创建移动任务成功", taskID)
}
