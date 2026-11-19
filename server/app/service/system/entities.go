package system

import (
	"context"
	"server/app/util/datastore"
	"server/app/util/psutil"
	"sync"
)

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
