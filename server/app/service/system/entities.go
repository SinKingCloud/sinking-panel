package system

import (
	"context"
	"sync"
	"time"

	"server/app/util/psutil"
	"server/app/util/recordstore"
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

// monitorSample 保存前端增量读取所需的短期 I/O 快照，不写入资源工具类。
type monitorSample struct {
	ID      int64                       `json:"sample_id"`
	Time    int64                       `json:"sample_time"`
	Network map[string]psutil.NetworkIO `json:"network"`
	Disk    map[string]psutil.DiskIO    `json:"disk"`
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
	Time    int64                `json:"time"`
	Items   []monitorProcessItem `json:"items"`
	Partial bool                 `json:"partial,omitempty"`
}

type monitorProcessData struct {
	CPU    *monitorProcessRanking `json:"cpu,omitempty"`
	Memory *monitorProcessRanking `json:"memory,omitempty"`
	DiskIO *monitorProcessRanking `json:"disk_io,omitempty"`
}

// MonitorProcessGroup 是区间内 Top 5 合计占用最高的一次采样，而非曲线均值时刻。
// Time 保留真实采样时间；无记录、权限不足和零占用有不同的可用性标记。
type MonitorProcessGroup struct {
	Time      int64                   `json:"time"`
	Items     []psutil.ProcessTopItem `json:"items"`
	Available bool                    `json:"available"`
	Partial   bool                    `json:"partial"`
}

type processPeaks [3]MonitorProcessGroup

type pendingProcessPeaks struct {
	peaks   processPeaks
	version uint64
}

// monitorProcessStore 只接受有界 Top 5 快照。采集线程仅合并内存，文件 I/O 全部交给单个 worker。
type monitorProcessStore struct {
	store      *recordstore.Store
	collection *recordstore.Collection
	path       string
	cancel     context.CancelFunc
	done       chan struct{}
	wake       chan struct{}
	queries    chan struct{}
	mu         sync.Mutex
	pending    map[int64]pendingProcessPeaks
	enabled    bool
	closed     bool
	days       int
	acceptFrom int64
	lastError  error
	dropped    uint64
	closeError error
}
