package task

import "time"

// Create 创建任务
func (s *Service) Create(id, name string, data interface{}) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	task := &Task{
		ID:         id,
		Name:       name,
		Status:     StatusPending,
		Progress:   0,
		Data:       data,
		CreateTime: now,
		UpdateTime: now,
	}
	s.tasks[id] = task
	return task
}
