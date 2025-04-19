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

func Compress(c *server.Context) {
	type Form struct {
		Paths  []string `json:"paths" default:"" validate:"required" label:"文件路径"`
		Dir    string   `json:"dir" default:"" validate:"required" label:"目标目录"`
		Format string   `json:"format" default:"zip" validate:"required" label:"压缩格式"`
		Name   string   `json:"name" default:"" validate:"required" label:"文件名"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	if !service.File.IsSupportedFormat(form.Format) {
		c.Error("不支持的压缩格式")
		return
	}
	f := file.NewDisk("")
	for _, path := range form.Paths {
		if !f.Exists(path) {
			c.Error("源文件或目录不存在: " + path)
			return
		}
	}
	if !f.Exists(form.Dir) {
		err := f.CreateDir(form.Dir)
		if err != nil {
			c.Error("创建目标目录失败: " + err.Error())
			return
		}
	}
	destPath := form.Dir + "/" + form.Name
	if form.Format != "" {
		if form.Format == "tgz" {
			if !service.File.IsSupportedFormat("tgz") {
				destPath += ".tar.gz"
			} else {
				destPath += ".tgz"
			}
		} else {
			destPath += "." + form.Format
		}
	}
	if f.Exists(destPath) {
		c.Error("目标文件已存在: " + destPath)
		return
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	taskName := "压缩文件: " + form.Name + "." + form.Format
	taskData := map[string]interface{}{
		"paths":     form.Paths,
		"dest_path": destPath,
		"format":    form.Format,
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
		service.System.TaskUpdate(taskID, system.TaskStatusRunning, 0, "开始压缩")
		err := service.File.CompressFiles(taskInfo.Context, form.Paths, destPath, form.Format, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if canceled.Load() {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			message := "正在压缩: " + currentFile
			service.System.TaskUpdate(taskID, system.TaskStatusRunning, progress, message)
			return true
		})
		select {
		case <-taskInfo.Context.Done():
			service.System.TaskUpdate(taskID, system.TaskStatusCanceled, 0, "任务已取消")
			_ = f.Delete(destPath)
			return
		default:
			if canceled.Load() {
				service.System.TaskUpdate(taskID, system.TaskStatusCanceled, 0, "任务已取消")
				_ = f.Delete(destPath)
				return
			}
			if err != nil {
				service.System.TaskUpdate(taskID, system.TaskStatusFailed, 0, "压缩失败: "+err.Error())
				_ = f.Delete(destPath)
				return
			}
			service.System.TaskUpdate(taskID, system.TaskStatusCompleted, 100, "压缩完成")
		}
	}()
	c.SuccessWithData("创建压缩任务成功", taskID)
}
