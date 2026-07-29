package system

import (
	"context"
)

// SetTaskCancelFunc 设置取消函数
func (s *service) SetTaskCancelFunc(id string, cancelFunc func()) {
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
		cancel()
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}
