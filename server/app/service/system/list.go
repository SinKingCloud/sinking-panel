package system

// TaskList 获取任务列表
func (s *service) TaskList() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		snapshot := *task
		tasks = append(tasks, &snapshot)
	}
	return tasks
}
