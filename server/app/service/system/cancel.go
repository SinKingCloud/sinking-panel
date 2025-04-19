package system

import "time"

// TaskCancel 取消任务
func (s *Service) TaskCancel(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return false
	}
	// 只有运行中的任务才能取消
	if task.Status != TaskStatusRunning {
		return false
	}
	// 执行取消函数
	if task.CancelFunc != nil {
		task.CancelFunc()
	}
	// 更新任务状态
	task.Status = TaskStatusCanceled
	task.Message = "任务已取消"
	task.UpdateTime = time.Now().Unix()
	task.EndTime = task.UpdateTime
	return true
}
