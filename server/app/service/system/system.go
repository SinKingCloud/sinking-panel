package system

import (
	"context"
	"log"
	"os"
	"server/app/constant"
	serviceConfig "server/app/service/config"
	"server/app/util/container"
	"server/app/util/daemon"
	"server/app/util/psutil"
	"server/app/util/recordstore"
	"sync"
	"sync/atomic"
)

// Service 暴露方法。
type Service interface {
	Start(monitor ...bool)
	Close()
	Restart(action string, containers bool) error                            // 重启面板或服务器，可选同时重启容器
	GetInfo() map[string]interface{}                                         // 主机、系统和运行时信息
	GetTimezone() (map[string]interface{}, error)                            // 本机时区、时区列表和系统当地时间
	SetTimezone(name string) error                                           // 设置系统时区
	GetStatis() map[string]interface{}                                       // CPU、内存、磁盘等资源统计
	GetIO(netInterface, diskName string, after int64) map[string]interface{} // 网络与磁盘 I/O 统计
	GetMonitorSettings() (recordstore.SeriesSettings, error)
	SetMonitorSettings(enabled bool, days int) error
	GetMonitorHistory(start, end int64, maxPoints ...int) (recordstore.RecordResult, error)
	GetTask(id string) *Task
	TaskList(status *int) ([]*Task, error)
	TaskCreate(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string)))
	TaskUpdate(id string, status int, progress float64, message string, result ...interface{})
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
	configService          serviceConfig.Service
	daemon                 *daemon.Daemon
	container              *container.Manager
	restartPending         atomic.Bool
	maxTaskWorkers         int    // 系统任务并发 worker 数量
	systemTaskLogDirectory string // 系统任务日志目录

	queueMu   sync.Mutex
	queueCond *sync.Cond
	queue     []string
	jobs      map[string]*job
	logMu     sync.RWMutex

	monitorMu        sync.RWMutex
	resourceSampler  *psutil.Sampler
	systemBaseCache  psutil.HostInfo
	resourceCache    psutil.Snapshot
	disksInfoCache   []psutil.DiskInfo
	netRateCache     map[string]psutil.NetworkIO
	diskRateCache    map[string]psutil.DiskIO
	monitorSequence  int64
	monitorUpdatedAt int64
	monitorHistory   []monitorSample
	monitorRecorder  *recordstore.Recorder
	monitorInitError error
	monitorDisks     map[string]bool
	monitorNetworks  map[string]bool
	monitorDiskReady bool
}

// NewService 实例化service
func NewService(configService serviceConfig.Service, daemon *daemon.Daemon, container *container.Manager) *service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		ctx:                    ctx,
		cancel:                 cancel,
		tasks:                  make(map[string]*Task),
		jobs:                   make(map[string]*job),
		queue:                  make([]string, 0),
		configService:          configService,
		daemon:                 daemon,
		container:              container,
		maxTaskWorkers:         maxTaskWorkers,
		systemTaskLogDirectory: constant.TaskPath,
		resourceSampler:        psutil.NewSampler(),
		disksInfoCache:         make([]psutil.DiskInfo, 0),
		resourceCache:          psutil.Snapshot{Network: make([]psutil.NetworkIO, 0)},
		netRateCache:           make(map[string]psutil.NetworkIO),
		diskRateCache:          make(map[string]psutil.DiskIO),
		monitorHistory:         make([]monitorSample, 0, 120),
	}
	s.queueCond = sync.NewCond(&s.queueMu)
	return s
}

// Start 启动任务 worker；命令行传入 false 时保留日志且不启动系统监控。
func (s *service) Start(monitor ...bool) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.started || s.closed {
		return
	}
	background := len(monitor) == 0 || monitor[0]
	// 只有面板正常启动才清理上一次进程遗留的任务日志。
	if background {
		_ = os.RemoveAll(s.systemTaskLogDirectory)
	}
	_ = os.MkdirAll(s.systemTaskLogDirectory, 0755)
	s.started = true
	s.wait.Add(s.maxTaskWorkers)
	for i := 0; i < s.maxTaskWorkers; i++ {
		go s.taskWorker()
	}
	if background {
		s.startMonitor()
	}
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
	if s.monitorRecorder != nil {
		if err := s.monitorRecorder.Close(); err != nil {
			log.Printf("保存系统监控历史失败: %v", err)
		}
	}
}
