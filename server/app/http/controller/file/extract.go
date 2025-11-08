package file

import (
	"server/app/service"
	"server/app/service/system"
	"server/app/util/context"
	"server/app/util/file"
	"server/app/util/str"
	"sync/atomic"
	"time"
)

func Extract(c *context.Context) {
	type Form struct {
		Path string `json:"path" default:"" validate:"required" label:"压缩文件路径"`
		Dir  string `json:"dir" default:"" validate:"required" label:"解压目标目录"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.Path) {
		c.Error("压缩文件不存在")
		return
	}
	if f.IsDir(form.Path) {
		c.Error("不能解压目录，请选择压缩文件")
		return
	}
	format := service.File.GetFormatByExt(form.Path)
	if !service.File.IsSupportedFormat(format) {
		c.Error("不支持的压缩格式或不是有效的压缩文件")
		return
	}
	if !f.Exists(form.Dir) {
		err := f.CreateDir(form.Dir)
		if err != nil {
			c.Error("创建目标目录失败: " + err.Error())
			return
		}
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	taskName := "解压文件: " + form.Path
	taskData := map[string]interface{}{
		"path":     form.Path,
		"dest_dir": form.Dir,
		"format":   format,
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
		service.System.TaskUpdate(taskID, system.TaskStatusRunning, 0, "开始解压")
		err := service.File.Extract(taskInfo.Context, form.Path, form.Dir, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if canceled.Load() {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			message := "正在解压: " + currentFile
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
				service.System.TaskUpdate(taskID, system.TaskStatusFailed, 0, "解压失败: "+err.Error())
				return
			}
			service.System.TaskUpdate(taskID, system.TaskStatusCompleted, 100, "解压完成")
		}
	}()
	c.SuccessWithData("创建解压任务成功", taskID)
}
