package system

import (
	"context"
	"errors"
	"server/app/enum/system_task_status"
)

// NewTaskGroup 创建独立任务组，业务服务销毁时调用 Close 取消自己的任务。
func NewTaskGroup(system Service) *TaskGroup {
	return &TaskGroup{
		system: system,
		jobs:   make(map[string]context.CancelFunc),
	}
}

// Create 将任务提交到系统队列，只跟踪本组任务，不额外创建后台协程。
func (g *TaskGroup) Create(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string))) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return errors.New("业务任务组已关闭")
	}
	if g.system == nil {
		return errors.New("系统任务服务未初始化")
	}
	if run == nil {
		return errors.New("任务执行方法不能为空")
	}
	// 外部取消的排队任务不会进入回调，提交新任务时顺便清理这些记录。
	for id, cancel := range g.jobs {
		if cancel != nil {
			continue
		}
		task := g.system.GetTask(id)
		if task == nil || (task.Status != system_task_status.Pending && task.Status != system_task_status.Running) {
			delete(g.jobs, id)
		}
	}
	if _, exists := g.jobs[id]; exists {
		return errors.New("任务ID已存在")
	}
	g.jobs[id] = nil
	err := g.system.TaskCreate(id, name, data, func(ctx context.Context, data interface{}, update func(int, float64, string)) {
		g.mu.Lock()
		if g.closed {
			g.mu.Unlock()
			update(system_task_status.Canceled, 0, "任务已取消")
			return
		}
		if ctx.Err() != nil {
			g.mu.Unlock()
			return
		}
		ctx, cancel := context.WithCancel(ctx)
		g.jobs[id] = cancel
		g.wait.Add(1)
		g.mu.Unlock()
		defer func() {
			cancel()
			g.mu.Lock()
			delete(g.jobs, id)
			g.mu.Unlock()
			g.wait.Done()
		}()
		run(ctx, data, update)
	})
	if err != nil {
		delete(g.jobs, id)
	}
	return err
}

// Close 取消本组排队和执行中的任务，仅等待已开始的回调完成，可重复调用。
func (g *TaskGroup) Close() {
	g.closeOnce.Do(func() {
		g.mu.Lock()
		g.closed = true
		jobs := g.jobs
		g.jobs = make(map[string]context.CancelFunc)
		g.mu.Unlock()
		for id, cancel := range jobs {
			if cancel != nil {
				cancel()
			}
			g.system.TaskCancel(id)
		}
		g.wait.Wait()
	})
}
