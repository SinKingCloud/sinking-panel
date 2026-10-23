package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"server/app/enum/database_type"
	"server/app/enum/system_task_status"
	"server/app/util/str"
	"strings"
	"time"
)

// Transfer 复用系统任务队列；连接与文件处理在 worker 内执行，不依赖 HTTP 请求生命周期。
func (s *service) Transfer(id int64, databaseType int, action, path, databaseName string, run func(context.Context, func(float64, string)) (map[string]interface{}, error)) (string, error) {
	label := "导入"
	if action == "export" {
		label = "导出"
	} else if action != "import" {
		return "", errors.New("导入导出操作不合法")
	}
	if databaseName == "" || len(databaseName) > 1024 || strings.ContainsRune(databaseName, 0) {
		return "", errors.New("请选择数据库")
	}
	if run == nil || s.systemService == nil {
		return "", errors.New("后台任务服务不可用")
	}
	if !filepath.IsAbs(path) {
		return "", errors.New("请填写服务器上的完整文件路径")
	}
	if action == "import" {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() {
			return "", errors.New("导入路径必须是普通文件")
		}
	}
	data, err := s.FindConnection(id, databaseType)
	if err != nil {
		return "", err
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	name := label + " " + database_type.Map()[databaseType] + ": " + data.Name + "/" + databaseName
	params := map[string]interface{}{"id": id, "type": databaseType, "action": action, "path": path, "database_name": databaseName}
	s.systemService.TaskCreate(taskID, name, params, func(ctx context.Context, _ interface{}, update func(int, float64, string)) {
		percent := float64(0)
		lastUpdate := time.Time{}
		update(system_task_status.Running, 0, "正在准备"+label)
		result, err := run(ctx, func(current float64, message string) {
			if ctx.Err() != nil || math.IsNaN(current) || math.IsInf(current, 0) {
				return
			}
			current = math.Max(percent, math.Min(99, current))
			now := time.Now()
			if now.Sub(lastUpdate) < 250*time.Millisecond && current-percent < 1 {
				return
			}
			percent, lastUpdate = current, now
			update(system_task_status.Running, percent, message)
		})
		status, message := system_task_status.Completed, label+"完成"
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			status, message = system_task_status.Canceled, "任务已取消"
		} else if err != nil {
			status, message = system_task_status.Failed, label+"失败: "+err.Error()
		} else {
			percent = 100
		}
		if result != nil && result["success"] != nil {
			message += fmt.Sprintf("；已处理 %v %v", result["success"], result["unit"])
		}
		s.systemService.TaskUpdate(taskID, status, percent, message, result)
	})
	if s.systemService.GetTask(taskID) == nil {
		return "", errors.New("创建后台任务失败，面板正在停止")
	}
	return taskID, nil
}
