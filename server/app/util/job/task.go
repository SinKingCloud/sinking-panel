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

type taskRun struct {
	ctx      context.Context
	cancel   context.CancelFunc
	jobs     chan interface{}
	done     chan struct{}
	workers  sync.WaitGroup
	producer func(context.Context, chan interface{}) error
	consumer func(context.Context, interface{}) error
}

type Task struct {
	Thread   int
	Producer func(ctx context.Context, channel chan interface{}) error
	Consumer func(ctx context.Context, param interface{}) error

	status int32
	mutex  sync.RWMutex
	run    *taskRun
	stats  TaskStats
}

// NewTask 创建新任务
func NewTask() *Task {
	return &Task{Thread: 1}
}

// SetThread 设置协程数量
func (task *Task) SetThread(num int) *Task {
	if num <= 0 {
		num = 1
	}
	task.mutex.Lock()
	task.Thread = num
	task.mutex.Unlock()
	return task
}

// SetProducer 设置生产者
func (task *Task) SetProducer(fun func(ctx context.Context, channel chan interface{}) error) *Task {
	task.mutex.Lock()
	task.Producer = fun
	task.mutex.Unlock()
	return task
}

// SetConsumer 设置消费者
func (task *Task) SetConsumer(fun func(ctx context.Context, param interface{}) error) *Task {
	task.mutex.Lock()
	task.Consumer = fun
	task.mutex.Unlock()
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
	return task.stats
}

// IsRunning 检查任务是否正在运行
func (task *Task) IsRunning() bool {
	status := task.GetStatus()
	return status == TaskRunning || status == TaskStopping
}

// Start 启动任务
func (task *Task) Start() error {
	return task.StartWithContext(context.Background())
}

// StartWithContext 使用指定上下文启动任务
func (task *Task) StartWithContext(ctx context.Context) error {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	if task.IsRunning() {
		return fmt.Errorf("任务已经在运行中，当前状态: %v", task.GetStatus())
	}
	threads := task.Thread
	if threads <= 0 {
		threads = 1
	}
	if ctx == nil {
		ctx = context.Background()
	}
	run := &taskRun{
		jobs:     make(chan interface{}, threads*2),
		done:     make(chan struct{}),
		producer: task.Producer,
		consumer: task.Consumer,
	}
	run.ctx, run.cancel = context.WithCancel(ctx)
	task.run = run
	task.stats = TaskStats{StartTime: time.Now()}
	atomic.StoreInt32(&task.status, int32(TaskRunning))
	run.workers.Add(threads)
	for i := 0; i < threads; i++ {
		go task.worker(run)
	}
	go task.produce(run)
	return nil
}

func (task *Task) worker(run *taskRun) {
	task.mutex.Lock()
	task.stats.ActiveWorkers++
	task.mutex.Unlock()
	defer func() {
		task.mutex.Lock()
		task.stats.ActiveWorkers--
		task.mutex.Unlock()
		run.workers.Done()
	}()
	for value := range run.jobs {
		if run.ctx.Err() != nil || run.consumer == nil {
			continue
		}
		err := run.consumer(run.ctx, value)
		task.mutex.Lock()
		if err != nil {
			task.stats.FailedJobs++
		} else {
			task.stats.ProcessedJobs++
		}
		task.mutex.Unlock()
	}
}

func (task *Task) produce(run *taskRun) {
	if run.producer != nil {
		_ = run.producer(run.ctx, run.jobs)
	}
	close(run.jobs)
	run.workers.Wait()
	task.mutex.Lock()
	if task.run == run {
		task.stats.EndTime = time.Now()
		if task.GetStatus() == TaskStopping {
			atomic.StoreInt32(&task.status, int32(TaskStopped))
		} else {
			atomic.StoreInt32(&task.status, int32(TaskCompleted))
		}
	}
	task.mutex.Unlock()
	close(run.done)
}

// Stop 优雅停止任务
func (task *Task) Stop() error {
	return task.StopWithTimeout(30 * time.Second)
}

// StopWithTimeout 在指定时间内停止任务
func (task *Task) StopWithTimeout(timeout time.Duration) error {
	task.mutex.Lock()
	if !task.IsRunning() {
		task.mutex.Unlock()
		return nil
	}
	run := task.run
	atomic.StoreInt32(&task.status, int32(TaskStopping))
	run.cancel()
	task.mutex.Unlock()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-run.done:
		return nil
	case <-timer.C:
		return fmt.Errorf("停止任务超时 (%v)", timeout)
	}
}

// Wait 等待任务完成
func (task *Task) Wait() {
	task.mutex.RLock()
	run := task.run
	task.mutex.RUnlock()
	if run != nil {
		<-run.done
	}
}

// WaitWithTimeout 在指定时间内等待任务完成
func (task *Task) WaitWithTimeout(timeout time.Duration) error {
	task.mutex.RLock()
	run := task.run
	task.mutex.RUnlock()
	if run == nil {
		return nil
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-run.done:
		return nil
	case <-timer.C:
		return fmt.Errorf("等待任务完成超时 (%v)", timeout)
	}
}
