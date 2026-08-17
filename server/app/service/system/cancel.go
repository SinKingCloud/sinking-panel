package system

import (
	"server/app/enum/system_task_status"
	"time"
)

// TaskCancel 取消任务
func (s *service) TaskCancel(id string) bool {
	s.mu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	// 等待中和运行中的任务都可以取消。
	if task.Status != system_task_status.Pending && task.Status != system_task_status.Running {
		s.mu.Unlock()
		return false
	}
	cancel := task.cancel
	// 更新任务状态
	task.Status = system_task_status.Canceled
	task.Message = "任务已取消"
	task.UpdateTime = time.Now().Unix()
	task.EndTime = task.UpdateTime
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.queueMu.Lock()
	delete(s.jobs, id)
	s.queueMu.Unlock()
	return true
}
