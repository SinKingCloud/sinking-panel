package task

import (
	"context"
	"sync"
	"time"
)

// Task 任务信息
type Task struct {
	ID         string             `json:"id"`          // 任务ID
	Name       string             `json:"name"`        // 任务名称
	Status     Status             `json:"status"`      // 任务状态
	Progress   float64            `json:"progress"`    // 进度（0-100）
	Message    string             `json:"message"`     // 消息
	Data       interface{}        `json:"data"`        // 数据
	StartTime  int64              `json:"start_time"`  // 开始时间
	EndTime    int64              `json:"end_time"`    // 结束时间
	CreateTime int64              `json:"create_time"` // 创建时间
	UpdateTime int64              `json:"update_time"` // 更新时间
	CancelFunc func()             `json:"-"`           // 取消函数，不序列化
	Context    context.Context    `json:"-"`           // 上下文，不序列化
	Cancel     context.CancelFunc `json:"-"`           // 取消函数，不序列化
}

// Service 任务服务
type Service struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

var (
	instance *Service
	once     sync.Once
)

// GetIns 获取任务服务实例（单例模式）
func GetIns() *Service {
	once.Do(func() {
		instance = &Service{
			tasks: make(map[string]*Task),
		}
	})
	return instance
}

// CreateTask 创建任务
func (s *Service) CreateTask(id, name string, data interface{}) *Task {
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

// GetTask 获取任务
func (s *Service) GetTask(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.tasks[id]
}

// UpdateTask 更新任务
func (s *Service) UpdateTask(id string, status Status, progress float64, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return
	}

	task.Status = status
	task.Progress = progress
	task.Message = message
	task.UpdateTime = time.Now().Unix()

	if status == StatusRunning && task.StartTime == 0 {
		task.StartTime = task.UpdateTime
	}

	if status == StatusCompleted || status == StatusFailed || status == StatusCanceled {
		task.EndTime = task.UpdateTime
	}
}

// SetCancelFunc 设置取消函数
func (s *Service) SetCancelFunc(id string, cancelFunc func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return
	}

	// 创建带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())
	task.Context = ctx
	task.Cancel = cancel

	// 设置取消函数
	task.CancelFunc = func() {
		cancel()     // 取消上下文
		cancelFunc() // 执行用户提供的取消函数
	}
}

// CancelTask 取消任务
func (s *Service) CancelTask(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return false
	}

	// 只有运行中的任务才能取消
	if task.Status != StatusRunning {
		return false
	}

	// 执行取消函数
	if task.CancelFunc != nil {
		task.CancelFunc()
	}

	// 更新任务状态
	task.Status = StatusCanceled
	task.Message = "任务已取消"
	task.UpdateTime = time.Now().Unix()
	task.EndTime = task.UpdateTime

	return true
}

// ListTasks 获取任务列表
func (s *Service) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// DeleteTask 删除任务
func (s *Service) DeleteTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, id)
}
