package file

import (
	"server/app/service"
	"server/app/service/task"
	"server/app/util/file"
	"server/app/util/server"
	"server/app/util/str"
	"sync/atomic"
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
	_ = service.Task.Create(taskID, taskName, taskData)
	var canceled atomic.Bool
	service.Task.SetCancelFunc(taskID, func() {
		canceled.Store(true)
	})
	go func() {
		taskInfo := service.Task.Get(taskID)
		if taskInfo == nil || taskInfo.Context == nil {
			return
		}
		service.Task.Update(taskID, task.StatusRunning, 0, "开始移动")
		_, _, _, err := f.Count(form.SourcePath)
		if err != nil {
			service.Task.Update(taskID, task.StatusFailed, 0, "计算文件大小失败: "+err.Error())
			return
		}
		err = f.MoveWithProcess(form.SourcePath, form.TargetPath, func(current, total int64, currentFile string, totalFiles, currentIndex int64) {
			select {
			case <-taskInfo.Context.Done():
				return
			default:
				if canceled.Load() {
					return
				}
				progress := float64(0)
				if total > 0 {
					progress = float64(current) / float64(total) * 100
				}
				message := "正在移动: " + currentFile
				service.Task.Update(taskID, task.StatusRunning, progress, message)
			}
		})
		select {
		case <-taskInfo.Context.Done():
			service.Task.Update(taskID, task.StatusCanceled, 0, "任务已取消")
			return
		default:
			if canceled.Load() {
				service.Task.Update(taskID, task.StatusCanceled, 0, "任务已取消")
				return
			}
			if err != nil {
				service.Task.Update(taskID, task.StatusFailed, 0, "移动失败: "+err.Error())
				return
			}
			service.Task.Update(taskID, task.StatusCompleted, 100, "移动完成")
		}
	}()
	c.SuccessWithData("创建任务成功", taskID)
}
