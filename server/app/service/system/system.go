package system

import (
	"server/app/service/file"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/net"
)

// Service service接口
type Service interface {
	GetInfo() map[string]interface{}
	GetStatus(netInterface, diskName string) map[string]interface{}
	GetTask(id string) *Task
	TaskList() []*Task
	TaskCreate(id, name string, data interface{}) *Task
	TaskDelete(id string)
	TaskUpdate(id string, status int, progress float64, message string)
	TaskCancel(id string) bool
	SetTaskCancelFunc(id string, cancelFunc func())
}

// service 注入结构
type service struct {
	tasks       map[string]*Task
	mu          sync.RWMutex
	fileService file.Service

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
	netRateCache         map[string]map[string]interface{}
	netRateCacheLock     sync.RWMutex
	diskRateCache        map[string]map[string]interface{}
	diskRateCacheLock    sync.RWMutex
	lastUpdateTime       time.Time
}

// NewService 实例化service
func NewService(fileService file.Service) *service {
	s := &service{
		tasks:            make(map[string]*Task),
		fileService:      fileService,
		systemBaseCache:  make(map[string]interface{}),
		cpuInfoCache:     make(map[string]interface{}),
		memoryInfoCache:  make(map[string]interface{}),
		disksInfoCache:   make([]file.Disk, 0),
		loadInfoCache:    make(map[string]interface{}),
		runtimeInfoCache: make(map[string]interface{}),
		networkInfoCache: make([]map[string]interface{}, 0),
		netIOCache:       make(map[string]net.IOCountersStat),
		diskIOCache:      make(map[string]disk.IOCountersStat),
		netRateCache:     make(map[string]map[string]interface{}),
		diskRateCache:    make(map[string]map[string]interface{}),
		lastUpdateTime:   time.Now(),
	}
	s.startMonitor()
	return s
}
