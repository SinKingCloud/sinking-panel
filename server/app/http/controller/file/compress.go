package file

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/enum/system_task_status"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"server/app/util/str"
)

func Compress(c *context.Context) {
	var form struct {
		Paths  []string `json:"paths" default:"" validate:"required" label:"文件路径"`
		Dir    string   `json:"dir" default:"" validate:"required" label:"目标目录"`
		Format string   `json:"format" default:"zip" validate:"required" label:"压缩格式"`
		Name   string   `json:"name" default:"" validate:"required" label:"文件名"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
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
		"paths":     append([]string(nil), form.Paths...),
		"dest_path": destPath,
		"format":    form.Format,
	}
	service.System.TaskCreate(taskID, taskName, taskData, func(ctx stdContext.Context, rawData interface{}, update func(int, float64, string)) {
		data, ok := rawData.(map[string]interface{})
		if !ok {
			update(system_task_status.Failed, 0, "压缩任务数据无效")
			return
		}
		paths, pathsOK := data["paths"].([]string)
		destPath, destPathOK := data["dest_path"].(string)
		format, formatOK := data["format"].(string)
		if !pathsOK || !destPathOK || !formatOK || len(paths) == 0 || destPath == "" || format == "" {
			update(system_task_status.Failed, 0, "压缩任务数据无效")
			return
		}
		disk := file.NewDisk("")
		update(system_task_status.Running, 0, "开始压缩")
		err := service.File.CompressFiles(ctx, paths, destPath, format, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if ctx.Err() != nil {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			message := "正在压缩: " + currentFile
			update(system_task_status.Running, progress, message)
			return true
		})
		select {
		case <-ctx.Done():
			update(system_task_status.Canceled, 0, "任务已取消")
			_ = disk.Delete(destPath)
			return
		default:
			if err != nil {
				update(system_task_status.Failed, 0, "压缩失败: "+err.Error())
				_ = disk.Delete(destPath)
				return
			}
			update(system_task_status.Completed, 100, "压缩完成")
		}
	})
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建压缩任务", taskName)
	c.SuccessWithData("创建压缩任务成功", taskID)
}
