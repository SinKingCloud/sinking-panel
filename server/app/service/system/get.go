package system

// GetTask 获取任务
func (s *service) GetTask(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task := s.tasks[id]
	if task == nil {
		return nil
	}
	snapshot := *task
	return &snapshot
}
