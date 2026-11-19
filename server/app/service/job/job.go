package job

import (
	"context"
	"os"
	"server/app/constant"
	"server/app/enum/system_task_status"
	"sync"
)

// Service 管理即时提交的后台任务及其运行日志。
type Service interface {
	Start(clearLogs bool)
	Close()
	Info(id string) *Job
	List(status *int) ([]*Job, error)
	Create(id string, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string))) error
	Update(id string, status int, progress float64, message string, result ...interface{})
	Cancel(id string) bool
	Delete(id string) bool
	Log(id string, after, before int64, pageSize int) (map[string]interface{}, error)
}

type service struct {
	ctx         context.Context
	cancel      context.CancelFunc
	wait        sync.WaitGroup
	closeOnce   sync.Once
	lifecycleMu sync.Mutex
	started     bool
	closed      bool

	mu        sync.RWMutex
	records   map[string]*Job
	queueMu   sync.Mutex
	queueCond *sync.Cond
	queue     []string
	jobs      map[string]*queuedJob

	logMu        sync.RWMutex
	logDirectory string
	maxWorkers   int
}

// NewService 创建后台任务服务，Start 启动 worker。
func NewService(maxWorkers int) *service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		ctx:          ctx,
		cancel:       cancel,
		records:      make(map[string]*Job),
		jobs:         make(map[string]*queuedJob),
		logDirectory: constant.TaskPath,
		maxWorkers:   maxWorkers,
	}
	s.queueCond = sync.NewCond(&s.queueMu)
	return s
}

// Start 启动任务 worker；面板启动时清理遗留日志，命令行执行时保留日志。
func (s *service) Start(clearLogs bool) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.started || s.closed {
		return
	}
	if clearLogs {
		_ = os.RemoveAll(s.logDirectory)
	}
	_ = os.MkdirAll(s.logDirectory, 0755)
	s.started = true
	s.wait.Add(s.maxWorkers)
	for range s.maxWorkers {
		go s.worker()
	}
}

// Close 取消排队和运行中的任务，等待 worker 退出。
func (s *service) Close() {
	s.closeOnce.Do(func() {
		s.lifecycleMu.Lock()
		s.closed = true
		s.cancel()
		s.lifecycleMu.Unlock()

		s.mu.RLock()
		ids := make([]string, 0, len(s.records))
		for id, record := range s.records {
			if record.Status == system_task_status.Pending || record.Status == system_task_status.Running {
				ids = append(ids, id)
			}
		}
		s.mu.RUnlock()
		for _, id := range ids {
			s.Cancel(id)
		}

		s.queueMu.Lock()
		clear(s.jobs)
		s.queue = nil
		s.queueCond.Broadcast()
		s.queueMu.Unlock()
		s.wait.Wait()
	})
}
