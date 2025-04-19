package system

import "time"

// TaskUpdate 更新任务
func (s *Service) TaskUpdate(id string, status TaskStatus, progress float64, message string) {
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

	if status == TaskStatusRunning && task.StartTime == 0 {
		task.StartTime = task.UpdateTime
	}

	if status == TaskStatusCompleted || status == TaskStatusFailed || status == TaskStatusCanceled {
		task.EndTime = task.UpdateTime
	}
}
