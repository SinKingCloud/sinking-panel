package file

import (
	"fmt"
	"path/filepath"
	"server/app/service"
	"server/app/service/task"
	"server/app/util/file"
	"server/app/util/server"
	"server/app/util/str"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Download 下载远程文件到本地
func Download(c *server.Context) {
	type Form struct {
		URL  string `json:"url" default:"" validate:"required,url" label:"远程文件URL"`
		Path string `json:"path" default:"" validate:"required" label:"目标路径"`
		Name string `json:"name" default:"" validate:"omitempty" label:"文件名"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}

	// 验证目标目录是否存在
	f := file.NewDisk("")
	if !f.IsDir(form.Path) {
		if !f.Exists(form.Path) {
			// 尝试创建目录
			err := f.CreateDir(form.Path)
			if err != nil {
				c.Error("目标目录不存在且无法创建: " + err.Error())
				return
			}
		} else {
			c.Error("目标路径不是一个目录")
			return
		}
	}

	// 如果未指定文件名，从URL中提取
	fileName := form.Name
	if fileName == "" {
		urlPath := strings.Split(form.URL, "?")[0] // 去除查询参数
		fileName = filepath.Base(urlPath)
		if fileName == "" || fileName == "." {
			fileName = "download_" + strconv.FormatInt(time.Now().Unix(), 10)
		}
	}

	// 完整的目标文件路径
	targetFilePath := filepath.Join(form.Path, fileName)

	// 创建下载任务
	taskID := str.GetSnowWorkIns().GetUuid()
	taskName := "下载文件: " + form.URL
	taskData := map[string]interface{}{
		"url":         form.URL,
		"target_path": targetFilePath,
	}
	_ = service.Task.Create(taskID, taskName, taskData)

	// 设置取消功能
	var canceled atomic.Bool
	service.Task.SetCancelFunc(taskID, func() {
		canceled.Store(true)
	})
	// 异步执行下载任务
	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			service.Task.Delete(taskID)
		}()
		taskInfo := service.Task.Get(taskID)
		if taskInfo == nil || taskInfo.Context == nil {
			return
		}
		service.Task.Update(taskID, task.StatusRunning, 0, "开始下载")
		err := service.File.DownloadWithProgress(taskInfo.Context, form.URL, targetFilePath, func(current, total int64, speed float64) bool {
			if canceled.Load() {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			speedStr := "未知"
			if speed > 0 {
				speedStr = service.File.FormatSize(int64(speed))
			}
			message := fmt.Sprintf("正在下载: %s / %s - %s/s",
				service.File.FormatSize(current),
				service.File.FormatSize(total),
				speedStr)

			service.Task.Update(taskID, task.StatusRunning, progress, message)
			return true
		})
		// 处理下载结果
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
				service.Task.Update(taskID, task.StatusFailed, 0, "下载失败: "+err.Error())
				return
			}
			service.Task.Update(taskID, task.StatusCompleted, 100, "下载完成")
		}
	}()

	c.SuccessWithData("创建下载任务成功", taskID)
}
