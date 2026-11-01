package oss

import (
	"context"
	"errors"
	"server/app/enum/system_task_status"
	utilOss "server/app/util/oss"
	"server/app/util/str"
	"time"
)

// transferTask 仅负责传输任务的入队、进度和结果，文件读写由各操作独立实现。
func (s *service) transferTask(id int64, action, label, key, path string, run func(context.Context, utilOss.ProgressFunc) (int64, error)) (string, error) {
	if s.systemService == nil {
		return "", errors.New("后台任务服务不可用")
	}
	data, err := s.FindById(id)
	if err != nil {
		return "", err
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	params := map[string]interface{}{
		"type":   "oss",
		"id":     id,
		"action": action,
		"key":    key,
		"path":   path,
	}
	s.systemService.TaskCreate(taskID, label+" OSS: "+data.Name+"/"+key, params, func(ctx context.Context, _ interface{}, update func(int, float64, string)) {
		percent := float64(0)
		lastUpdate := time.Time{}
		message := "正在" + label + " OSS 文件"
		update(system_task_status.Running, 0, message)
		size, err := run(ctx, func(current, total int64) {
			if ctx.Err() != nil || total <= 0 {
				return
			}
			value := min(99, max(percent, float64(current)/float64(total)*100))
			now := time.Now()
			if value == percent || (now.Sub(lastUpdate) < 250*time.Millisecond && value-percent < 1) {
				return
			}
			percent, lastUpdate = value, now
			// 固定消息由系统任务按秒记录，避免每个数据块都写日志。
			update(system_task_status.Running, percent, message)
		})
		status, message := system_task_status.Completed, label+"完成"
		var result interface{}
		if err == nil {
			percent = 100
			result = map[string]interface{}{
				"id":     id,
				"action": action,
				"key":    key,
				"path":   path,
				"size":   size,
			}
			if ctx.Err() != nil {
				status, message = system_task_status.Canceled, "文件已写入，任务收到取消请求"
			}
		} else if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			status, message = system_task_status.Canceled, "任务已取消"
		} else {
			status, message = system_task_status.Failed, label+"失败: "+err.Error()
		}
		s.systemService.TaskUpdate(taskID, status, percent, message, result)
	})
	if s.systemService.GetTask(taskID) == nil {
		return "", errors.New("创建后台任务失败，面板正在停止")
	}
	return taskID, nil
}
