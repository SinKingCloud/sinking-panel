package system

import (
	"server/app/enum/system_task_status"
	"time"
)

// TaskUpdate 更新任务
func (s *service) TaskUpdate(id string, status int, progress float64, message string) {
	if _, ok := system_task_status.Map()[status]; !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return
	}

	task.Status = status
	task.Progress = progress
	task.Message = message
	task.UpdateTime = time.Now().Unix()

	if status == system_task_status.Running && task.StartTime == 0 {
		task.StartTime = task.UpdateTime
	}

	if status == system_task_status.Completed || status == system_task_status.Failed || status == system_task_status.Canceled {
		task.EndTime = task.UpdateTime
	}
}
