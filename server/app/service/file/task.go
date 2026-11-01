package file

import (
	"context"
	"errors"
	"server/app/enum/system_task_status"
	"server/app/util/str"
)

// createTask 统一创建文件任务，并记录运行进度、取消和完成状态。
func (s *service) createTask(action, name string, data map[string]interface{}, run func(context.Context, func(float64, string)) error) (string, error) {
	if s.systemService == nil {
		return "", errors.New("后台任务服务不可用")
	}
	id := str.GetSnowWorkIns().GetUuid()
	s.systemService.TaskCreate(id, name, data, func(ctx context.Context, _ interface{}, update func(int, float64, string)) {
		update(system_task_status.Running, 0, "开始"+action)
		err := run(ctx, func(progress float64, message string) {
			update(system_task_status.Running, progress, message)
		})
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			update(system_task_status.Canceled, 0, "任务已取消")
		} else if err != nil {
			update(system_task_status.Failed, 0, action+"失败: "+err.Error())
		} else {
			update(system_task_status.Completed, 100, action+"完成")
		}
	})
	if s.systemService.GetTask(id) == nil {
		return "", errors.New("创建后台任务失败，面板正在停止")
	}
	return id, nil
}
