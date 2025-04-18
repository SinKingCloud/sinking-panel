package task

// List 获取任务列表
func (s *Service) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
