package system

import (
	"context"
	"os"
	"server/app/constant"
	serviceFile "server/app/service/file"
	"server/app/util/daemon"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/net"
)

// Service 暴露方法。
type Service interface {
	Start()
	Close()
	Restart(action string, containers bool) error                            // 重启面板或服务器，可选同时重启容器
	GetInfo() map[string]interface{}                                         // 主机、系统和运行时信息
	GetTimezone() (map[string]interface{}, error)                            // 本机时区、时区列表和系统当地时间
	SetTimezone(name string) error                                           // 设置系统时区
	GetStatis() map[string]interface{}                                       // CPU、内存、磁盘等资源统计
	GetIO(netInterface, diskName string, after int64) map[string]interface{} // 网络与磁盘 I/O 统计
	GetTask(id string) *Task
	TaskList(status *int) ([]*Task, error)
	TaskCreate(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string)))
	TaskUpdate(id string, status int, progress float64, message string)
	TaskCancel(id string) bool
	TaskDelete(id string) bool
	TaskLog(id string, after int64, before int64, pageSize int) (map[string]interface{}, error)
}

// service 注入结构
type service struct {
	ctx                    context.Context
	cancel                 context.CancelFunc
	wait                   sync.WaitGroup
	lifecycleMu            sync.Mutex
	started                bool
	closed                 bool
	tasks                  map[string]*Task
	mu                     sync.RWMutex
	fileService            serviceFile.Service
	daemon                 *daemon.Daemon
	restartPending         atomic.Bool
	maxTaskWorkers         int    // 系统任务并发 worker 数量
	systemTaskLogDirectory string // 系统任务日志目录

	queueMu   sync.Mutex
	queueCond *sync.Cond
	queue     []string
	jobs      map[string]*job
	logMu     sync.RWMutex

	systemBaseCache      map[string]interface{}
	systemBaseCacheLock  sync.RWMutex
	cpuInfoCache         map[string]interface{}
	cpuInfoCacheLock     sync.RWMutex
	cpuModel             string
	cpuCores             int
	cpuInfoInitialized   bool
	memoryInfoCache      map[string]interface{}
	memoryInfoCacheLock  sync.RWMutex
	disksInfoCache       []serviceFile.Disk
	disksInfoCacheLock   sync.RWMutex
	loadInfoCache        map[string]interface{}
	loadInfoCacheLock    sync.RWMutex
	networkInfoCache     []map[string]interface{}
	networkInfoCacheLock sync.RWMutex
	netIOCache           map[string]net.IOCountersStat
	diskIOCache          map[string]disk.IOCountersStat
	ioCacheLock          sync.RWMutex
	netRateCache         map[string]map[string]interface{}
	diskRateCache        map[string]map[string]interface{}
	netLastUpdateTime    time.Time
	diskLastUpdateTime   time.Time
	monitorSequence      int64
	monitorUpdatedAt     int64
	monitorHistory       []map[string]interface{}
}

// NewService 实例化service
func NewService(fileService serviceFile.Service, daemon *daemon.Daemon) *service {
	now := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		ctx:                    ctx,
		cancel:                 cancel,
		tasks:                  make(map[string]*Task),
		jobs:                   make(map[string]*job),
		queue:                  make([]string, 0),
		fileService:            fileService,
		daemon:                 daemon,
		maxTaskWorkers:         maxTaskWorkers,
		systemTaskLogDirectory: constant.TaskPath,
		systemBaseCache:        make(map[string]interface{}),
		cpuInfoCache:           make(map[string]interface{}),
		memoryInfoCache:        make(map[string]interface{}),
		disksInfoCache:         make([]serviceFile.Disk, 0),
		loadInfoCache:          make(map[string]interface{}),
		networkInfoCache:       make([]map[string]interface{}, 0),
		netIOCache:             make(map[string]net.IOCountersStat),
		diskIOCache:            make(map[string]disk.IOCountersStat),
		netRateCache:           make(map[string]map[string]interface{}),
		diskRateCache:          make(map[string]map[string]interface{}),
		netLastUpdateTime:      now,
		diskLastUpdateTime:     now,
		monitorHistory:         make([]map[string]interface{}, 0, 120),
	}
	s.queueCond = sync.NewCond(&s.queueMu)
	return s
}

// Start 启动任务 worker 和系统监控。
func (s *service) Start() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.started || s.closed {
		return
	}
	// 系统任务仅保存在内存中，启动时清理上一次进程遗留的日志。
	_ = os.RemoveAll(s.systemTaskLogDirectory)
	_ = os.MkdirAll(s.systemTaskLogDirectory, 0755)
	s.started = true
	s.wait.Add(s.maxTaskWorkers)
	for i := 0; i < s.maxTaskWorkers; i++ {
		go s.taskWorker()
	}
	s.startMonitor()
}

// Close 停止后台任务并等待全部协程退出。
func (s *service) Close() {
	s.lifecycleMu.Lock()
	if !s.closed {
		s.closed = true
		s.cancel()
		s.queueMu.Lock()
		s.queueCond.Broadcast()
		s.queueMu.Unlock()
	}
	s.lifecycleMu.Unlock()
	s.wait.Wait()
}
