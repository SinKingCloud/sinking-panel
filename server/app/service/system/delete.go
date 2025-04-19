package system

// TaskDelete 删除任务
func (s *Service) TaskDelete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}
