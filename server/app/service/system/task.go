package system

import (
	"context"
	"sync"
)

// Task 任务信息
type Task struct {
	ID         string             `json:"id"`          // 任务ID
	Name       string             `json:"name"`        // 任务名称
	Status     TaskStatus         `json:"status"`      // 任务状态
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

// SetTaskCancelFunc 设置取消函数
func (s *Service) SetTaskCancelFunc(id string, cancelFunc func()) {
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
