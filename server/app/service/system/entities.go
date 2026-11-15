package system

import (
	"context"
	"sync"
	"time"

	"server/app/util/datastore"
	"server/app/util/psutil"
)

// Task 任务信息
type Task struct {
	ID         string      `json:"id"`               // 任务ID
	Name       string      `json:"name"`             // 任务名称
	Status     int         `json:"status"`           // 任务状态
	Progress   float64     `json:"progress"`         // 进度（0-100）
	Message    string      `json:"message"`          // 消息
	Data       interface{} `json:"data"`             // 数据
	Result     interface{} `json:"result,omitempty"` // 执行结果，仅在任务结束后写入
	StartTime  int64       `json:"start_time"`       // 开始时间
	EndTime    int64       `json:"end_time"`         // 结束时间
	CreateTime int64       `json:"create_time"`      // 创建时间
	UpdateTime int64       `json:"update_time"`      // 更新时间
	cancel     context.CancelFunc
	lastLogAt  time.Time
}

type job struct {
	id   string
	ctx  context.Context
	data interface{}
	run  func(context.Context, interface{}, func(int, float64, string))
}

// TaskGroup 管理一个业务服务创建的系统任务，不持有或关闭系统服务。
type TaskGroup struct {
	system    Service
	mu        sync.Mutex
	closed    bool
	jobs      map[string]context.CancelFunc
	wait      sync.WaitGroup
	closeOnce sync.Once
}

// monitorSample 保存前端增量读取所需的短期 I/O 快照，不写入资源工具类。
type monitorSample struct {
	ID      int64                       `json:"sample_id"`
	Time    int64                       `json:"sample_time"`
	Network map[string]psutil.NetworkIO `json:"network"`
	Disk    map[string]psutil.DiskIO    `json:"disk"`
	Process *monitorProcessRanking      `json:"process,omitempty"` // 本时刻最近完成的磁盘 I/O 进程 Top 5，携带实际采样窗口
}

// MonitorHistory 是监控页面的返回结构，进程排行随对应时间桶一同返回。
type MonitorHistory struct {
	Data         []map[string]interface{} `json:"data"`
	Interval     int64                    `json:"interval"`
	Start        int64                    `json:"start"`
	End          int64                    `json:"end"`
	ProcessError string                   `json:"process_error,omitempty"`
}

type monitorProcessItem struct {
	PID   int32   `json:"pid"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type monitorProcessRanking struct {
	Time     int64                `json:"time"`
	Interval int64                `json:"interval"` // CPU、I/O 平均值覆盖的毫秒数，内存为采样时刻的值
	Items    []monitorProcessItem `json:"items"`
	Partial  bool                 `json:"partial,omitempty"`
	Peak     *float64             `json:"peak,omitempty"`
	IOScope  string               `json:"io_scope,omitempty"`
}

type monitorProcessData struct {
	CPU    *monitorProcessRanking `json:"cpu,omitempty"`
	Memory *monitorProcessRanking `json:"memory,omitempty"`
	DiskIO *monitorProcessRanking `json:"disk_io,omitempty"`
}

// MonitorProcessGroup 保留整机采样高点对应的进程排行，CPU、I/O 使用同窗口均值。
// Peak 的单位依次为整机 CPU 百分比、已用内存百分比、磁盘读写字节/秒；nil 表示未采样。
type MonitorProcessGroup struct {
	Time      int64                   `json:"time"`
	Interval  int64                   `json:"interval"` // 进程实际采样间隔，不能与整机一秒峰值混用
	Items     []psutil.ProcessTopItem `json:"items"`
	Available bool                    `json:"available"`
	Partial   bool                    `json:"partial"`
	Peak      *float64                `json:"peak,omitempty"`
	IOScope   string                  `json:"io_scope,omitempty"` // process 表示进程总 I/O，空值为磁盘 I/O
}

type processPeaks [3]MonitorProcessGroup

type pendingProcessPeaks struct {
	peaks   processPeaks
	version uint64
}

// monitorProcessStore 只接受有界 Top 5 快照。采集线程仅合并内存，文件 I/O 全部交给单个 worker。
type monitorProcessStore struct {
	minutes     *datastore.Timeline
	hours       *datastore.Timeline
	cancel      context.CancelFunc
	done        chan struct{}
	wake        chan struct{}
	queries     chan struct{}
	mu          sync.Mutex
	pending     map[int64]pendingProcessPeaks
	enabled     bool
	closed      bool
	days        int
	acceptFrom  int64
	lastError   error
	dropped     uint64
	closeError  error
	checkedHour int64 // 仅后台写入协程使用，避免重复检查已结束小时。
}
