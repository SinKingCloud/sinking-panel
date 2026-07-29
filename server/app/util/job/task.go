package job

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskStatus 任务状态
type TaskStatus int32

const (
	TaskStopped   TaskStatus = iota // 已停止
	TaskRunning                     // 运行中
	TaskStopping                    // 停止中
	TaskCompleted                   // 已完成
)

// TaskStats 任务统计信息
type TaskStats struct {
	StartTime     time.Time // 开始时间
	EndTime       time.Time // 结束时间
	ProcessedJobs int64     // 已处理任务数
	FailedJobs    int64     // 失败任务数
	ActiveWorkers int32     // 活跃工作协程数
}

type Task struct {
	Thread   int
	Producer func(ctx context.Context, channel chan interface{}) error
	Consumer func(ctx context.Context, param interface{}) error

	// 内部状态管理
	status   int32              // 使用atomic操作的状态
	ctx      context.Context    // 上下文
	cancel   context.CancelFunc // 取消函数
	wg       *sync.WaitGroup    // 等待组
	taskChan chan interface{}   // 任务通道
	doneChan chan struct{}      // 完成通道
	stats    TaskStats          // 统计信息
	mutex    sync.RWMutex       // 读写锁
}

// NewTask 创建新任务
func NewTask() *Task {
	return &Task{
		Thread:   1,
		wg:       &sync.WaitGroup{},
		doneChan: make(chan struct{}, 1), // 改为缓冲通道，避免阻塞
	}
}

// SetThread 设置协程数量
func (task *Task) SetThread(num int) *Task {
	if num <= 0 {
		num = 1
	}
	task.Thread = num
	return task
}

// SetProducer 设置生产者
func (task *Task) SetProducer(fun func(ctx context.Context, channel chan interface{}) error) *Task {
	task.Producer = fun
	return task
}

// SetConsumer 设置消费者
func (task *Task) SetConsumer(fun func(ctx context.Context, param interface{}) error) *Task {
	task.Consumer = fun
	return task
}

// GetStatus 获取任务状态
func (task *Task) GetStatus() TaskStatus {
	return TaskStatus(atomic.LoadInt32(&task.status))
}

// GetStats 获取任务统计信息
func (task *Task) GetStats() TaskStats {
	task.mutex.RLock()
	defer task.mutex.RUnlock()

	stats := task.stats
	stats.ActiveWorkers = atomic.LoadInt32(&task.stats.ActiveWorkers)
	return stats
}

// IsRunning 检查任务是否正在运行
func (task *Task) IsRunning() bool {
	status := task.GetStatus()
	return status == TaskRunning || status == TaskStopping
}

// Start 启动任务（支持多次启动）
func (task *Task) Start() error {
	return task.StartWithContext(context.Background())
}

// StartWithContext 使用指定上下文启动任务
func (task *Task) StartWithContext(ctx context.Context) error {
	// 检查是否已经在运行
	for {
		currentStatus := atomic.LoadInt32(&task.status)
		if currentStatus == int32(TaskRunning) || currentStatus == int32(TaskStopping) {
			return fmt.Errorf("任务已经在运行中，当前状态: %v", TaskStatus(currentStatus))
		}
		if atomic.CompareAndSwapInt32(&task.status, currentStatus, int32(TaskRunning)) {
			break
		}
		// CAS失败，重试
	}

	// 创建新的上下文和取消函数
	task.ctx, task.cancel = context.WithCancel(ctx)
	task.taskChan = make(chan interface{}, task.Thread*2) // 缓冲通道
	task.wg = &sync.WaitGroup{}

	// 重置统计信息
	task.mutex.Lock()
	task.stats = TaskStats{
		StartTime: time.Now(),
	}
	task.mutex.Unlock()

	// 启动工作协程
	task.startWorkers()

	// 启动生产者
	go task.runProducer()

	return nil
}

// startWorkers 启动工作协程
func (task *Task) startWorkers() {
	for i := 0; i < task.Thread; i++ {
		task.wg.Add(1)
		go task.worker()
	}
}

// worker 工作协程
func (task *Task) worker() {
	defer task.wg.Done()
	atomic.AddInt32(&task.stats.ActiveWorkers, 1)
	defer atomic.AddInt32(&task.stats.ActiveWorkers, -1)

	for {
		select {
		case <-task.ctx.Done():
			return
		case job, ok := <-task.taskChan:
			if !ok {
				return
			}

			// 处理任务
			if task.Consumer != nil {
				if err := task.Consumer(task.ctx, job); err != nil {
					atomic.AddInt64(&task.stats.FailedJobs, 1)
				} else {
					atomic.AddInt64(&task.stats.ProcessedJobs, 1)
				}
			}
		}
	}
}

// runProducer 运行生产者
func (task *Task) runProducer() {
	defer func() {
		close(task.taskChan)
		// 等待所有工作协程完成
		task.wg.Wait()
		// 更新状态和统计信息
		atomic.StoreInt32(&task.status, int32(TaskCompleted))
		task.mutex.Lock()
		task.stats.EndTime = time.Now()
		task.mutex.Unlock()
		// 通知任务完成
		select {
		case task.doneChan <- struct{}{}:
		default:
		}
	}()

	if task.Producer != nil {
		task.Producer(task.ctx, task.taskChan)
	}
}

// Stop 优雅停止任务
func (task *Task) Stop() error {
	return task.StopWithTimeout(30 * time.Second)
}

// StopWithTimeout 在指定超时时间内停止任务
func (task *Task) StopWithTimeout(timeout time.Duration) error {
	// 检查当前状态
	currentStatus := task.GetStatus()
	if currentStatus == TaskStopped || currentStatus == TaskCompleted {
		return nil
	}

	if !atomic.CompareAndSwapInt32(&task.status, int32(TaskRunning), int32(TaskStopping)) {
		return fmt.Errorf("无法停止任务，当前状态: %v", currentStatus)
	}

	// 取消上下文，通知所有协程停止
	if task.cancel != nil {
		task.cancel()
	}

	// 等待任务完成或超时
	done := make(chan struct{})
	go func() {
		if task.wg != nil {
			task.wg.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
	}

	// 更新状态
	atomic.StoreInt32(&task.status, int32(TaskStopped))
	task.mutex.Lock()
	if task.stats.EndTime.IsZero() {
		task.stats.EndTime = time.Now()
	}
	task.mutex.Unlock()

	// 清理资源
	task.ctx = nil
	task.cancel = nil

	return nil
}

// Wait 等待任务完成
func (task *Task) Wait() {
	if !task.IsRunning() {
		return
	}
	<-task.doneChan
}

// WaitWithTimeout 在指定时间内等待任务完成
func (task *Task) WaitWithTimeout(timeout time.Duration) error {
	if !task.IsRunning() {
		return nil
	}

	select {
	case <-task.doneChan:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("等待任务完成超时 (%v)", timeout)
	}
}
