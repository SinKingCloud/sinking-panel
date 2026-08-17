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

func Extract(c *context.Context) {
	var form struct {
		Path string `json:"path" default:"" validate:"required" label:"压缩文件路径"`
		Dir  string `json:"dir" default:"" validate:"required" label:"解压目标目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
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
	service.System.TaskCreate(taskID, taskName, taskData, func(ctx stdContext.Context, rawData interface{}, update func(int, float64, string)) {
		data, ok := rawData.(map[string]interface{})
		if !ok {
			update(system_task_status.Failed, 0, "解压任务数据无效")
			return
		}
		path, pathOK := data["path"].(string)
		destDir, destDirOK := data["dest_dir"].(string)
		if !pathOK || !destDirOK || path == "" || destDir == "" {
			update(system_task_status.Failed, 0, "解压任务数据无效")
			return
		}
		update(system_task_status.Running, 0, "开始解压")
		err := service.File.Extract(ctx, path, destDir, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
			if ctx.Err() != nil {
				return false
			}
			progress := float64(0)
			if total > 0 {
				progress = float64(current) / float64(total) * 100
			}
			message := "正在解压: " + currentFile
			update(system_task_status.Running, progress, message)
			return true
		})
		select {
		case <-ctx.Done():
			update(system_task_status.Canceled, 0, "任务已取消")
			return
		default:
			if err != nil {
				update(system_task_status.Failed, 0, "解压失败: "+err.Error())
				return
			}
			update(system_task_status.Completed, 100, "解压完成")
		}
	})
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建解压任务", taskName)
	c.SuccessWithData("创建解压任务成功", taskID)
}
