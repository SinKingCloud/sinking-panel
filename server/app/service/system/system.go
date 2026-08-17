package system

import (
	"context"
	"server/app/service/file"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/net"
)

const (
	maxTaskWorkers = 5
	taskRetention  = 3 * time.Second
)

// Service 是系统信息和异步任务服务的公开契约。
type Service interface {
	GetInfo() map[string]interface{}
	GetStatus(netInterface, diskName string, after int64) map[string]interface{}
	GetTask(id string) *Task
	TaskList() []*Task
	TaskCreate(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string)))
	TaskUpdate(id string, status int, progress float64, message string)
	TaskCancel(id string) bool
}

// service 注入结构
type service struct {
	tasks       map[string]*Task
	mu          sync.RWMutex
	fileService file.Service

	queueMu   sync.Mutex
	queueCond *sync.Cond
	queue     []string
	jobs      map[string]*job

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
	s := &service{
		tasks:              make(map[string]*Task),
		jobs:               make(map[string]*job),
		queue:              make([]string, 0),
		fileService:        fileService,
		systemBaseCache:    make(map[string]interface{}),
		cpuInfoCache:       make(map[string]interface{}),
		memoryInfoCache:    make(map[string]interface{}),
		disksInfoCache:     make([]file.Disk, 0),
		loadInfoCache:      make(map[string]interface{}),
		runtimeInfoCache:   make(map[string]interface{}),
		networkInfoCache:   make([]map[string]interface{}, 0),
		netIOCache:         make(map[string]net.IOCountersStat),
		diskIOCache:        make(map[string]disk.IOCountersStat),
		netRateCache:       make(map[string]map[string]interface{}),
		diskRateCache:      make(map[string]map[string]interface{}),
		netLastUpdateTime:  now,
		diskLastUpdateTime: now,
		monitorHistory:     make([]map[string]interface{}, 0, 120),
	}
	s.queueCond = sync.NewCond(&s.queueMu)
	for i := 0; i < maxTaskWorkers; i++ {
		go s.taskWorker()
	}
	go s.taskCleanupWorker()
	s.startMonitor()
	return s
}
