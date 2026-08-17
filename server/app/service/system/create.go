package system

import (
	"context"
	"server/app/enum/system_task_status"
	"time"
)

// TaskCreate 创建任务
func (s *service) TaskCreate(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string))) {
	now := time.Now().Unix()
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		ID:         id,
		Name:       name,
		Status:     system_task_status.Pending,
		Progress:   0,
		Data:       data,
		CreateTime: now,
		UpdateTime: now,
		cancel:     cancel,
	}
	s.mu.Lock()
	s.tasks[id] = task
	s.mu.Unlock()

	s.queueMu.Lock()
	s.jobs[id] = &job{id: id, ctx: ctx, data: data, run: run}
	s.queue = append(s.queue, id)
	s.queueCond.Signal()
	s.queueMu.Unlock()
}
