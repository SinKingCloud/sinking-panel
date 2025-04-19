package system

import "time"

// TaskCreate 创建任务
func (s *Service) TaskCreate(id, name string, data interface{}) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	task := &Task{
		ID:         id,
		Name:       name,
		Status:     TaskStatusPending,
		Progress:   0,
		Data:       data,
		CreateTime: now,
		UpdateTime: now,
	}
	s.tasks[id] = task
	return task
}
