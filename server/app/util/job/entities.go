package job

import (
	"context"
	"sync"
	"time"
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
