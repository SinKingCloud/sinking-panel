package task

import "time"

// Update 更新任务
func (s *Service) Update(id string, status Status, progress float64, message string) {
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

	if status == StatusRunning && task.StartTime == 0 {
		task.StartTime = task.UpdateTime
	}

	if status == StatusCompleted || status == StatusFailed || status == StatusCanceled {
		task.EndTime = task.UpdateTime
	}
}
