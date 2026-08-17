package system

// TaskDelete cancels an active task or removes a completed task and its log.
func (s *service) TaskDelete(id string) bool {
	// Always cancel first. This invalidates the worker context and removes a
	// queued job before the task record and its log are deleted.
	if s.TaskCancel(id) {
		s.mu.Lock()
		delete(s.tasks, id)
		s.mu.Unlock()
		s.removeTaskLog(id)
		return true
	}

	s.mu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	cancel := task.cancel
	delete(s.tasks, id)
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	s.queueMu.Lock()
	delete(s.jobs, id)
	s.queueMu.Unlock()
	s.removeTaskLog(id)
	return true
}
