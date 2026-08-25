package system

import (
	"context"
	"server/app/service/file"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/net"
)

// Service 暴露方法。
type Service interface {
	Start()
	Close()
	GetInfo() map[string]interface{}
	GetStatus(netInterface, diskName string, after int64) map[string]interface{}
	GetTask(id string) *Task
	TaskList() []*Task
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
	fileService            file.Service
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
	disksInfoCache       []file.Disk
	disksInfoCacheLock   sync.RWMutex
	loadInfoCache        map[string]interface{}
	loadInfoCacheLock    sync.RWMutex
	runtimeInfoCache     map[string]interface{}
	runtimeInfoCacheLock sync.RWMutex
	networkInfoCache     []map[string]interface{}
	networkInfoCacheLock sync.RWMutex
	netIOCache           map[string]net.IOCountersStat
	diskIOCache          map[string]disk.IOCountersStat
	statusCacheLock      sync.RWMutex
	netRateCache         map[string]map[string]interface{}
	diskRateCache        map[string]map[string]interface{}
	netLastUpdateTime    time.Time
	diskLastUpdateTime   time.Time
	monitorSequence      int64
	monitorUpdatedAt     int64
	monitorHistory       []map[string]interface{}
}

// NewService 实例化service
func NewService(fileService file.Service) *service {
	now := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		ctx:                    ctx,
		cancel:                 cancel,
		tasks:                  make(map[string]*Task),
		jobs:                   make(map[string]*job),
		queue:                  make([]string, 0),
		fileService:            fileService,
		maxTaskWorkers:         maxTaskWorkers,
		systemTaskLogDirectory: systemTaskLogDirectory,
		systemBaseCache:        make(map[string]interface{}),
		cpuInfoCache:           make(map[string]interface{}),
		memoryInfoCache:        make(map[string]interface{}),
		disksInfoCache:         make([]file.Disk, 0),
		loadInfoCache:          make(map[string]interface{}),
		runtimeInfoCache:       make(map[string]interface{}),
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
	s.resetTaskLogs()
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
