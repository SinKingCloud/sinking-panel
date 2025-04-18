package task

// Delete 删除任务
func (s *Service) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}
