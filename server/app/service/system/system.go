package system

import (
	"context"
	"os"
	"server/app/constant"
	"server/app/enum/system_task_status"
	serviceConfig "server/app/service/config"
	"server/app/util/container"
	"server/app/util/daemon"
	"server/app/util/datastore"
	"server/app/util/psutil"
	"sync"
	"sync/atomic"
	"time"
)

// Service 暴露方法。
type Service interface {
	Start(monitor ...bool)
	Close()
	Restart(action string, containers bool) error                            // 重启面板或服务器，可选同时重启容器
	GetInfo() map[string]interface{}                                         // 主机、系统和运行时信息
	GetTimezone() (map[string]interface{}, error)                            // 本机时区、时区列表和系统当地时间
	SetTimezone(name string) error                                           // 设置系统时区
	SyncTimezone(ctx context.Context) error                                  // 同步服务器时间
	GetStatis() map[string]interface{}                                       // CPU、内存、磁盘等资源统计
	GetIO(netInterface, diskName string, after int64) map[string]interface{} // 网络与磁盘 I/O 统计
	GetMonitorSettings() (datastore.SeriesSettings, error)
	SetMonitorSettings(enabled bool, days int) error
	GetMonitorHistory(ctx context.Context, start, end int64, maxPoints int) (MonitorHistory, error)
	GetTask(id string) *Task
	TaskList(status *int) ([]*Task, error)
	TaskCreate(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string))) error
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
	closeOnce              sync.Once
	lifecycleMu            sync.Mutex
	started                bool
	closed                 bool
	panelStartedAt         time.Time
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

	monitorMu             sync.RWMutex
	resourceSampler       *psutil.Sampler
	processSampler        *psutil.ProcessSampler
	processRequestedUntil atomic.Int64 // 实时接口只延长采集窗口，不触发进程扫描
	processSampleAt       time.Time    // 以下基线只由监控采集协程使用
	processBaseline       psutil.Snapshot
	processCache          psutil.ProcessTopSnapshot
	processIOCache        *monitorProcessRanking
	monitorProcessStore   *monitorProcessStore
	monitorProcessError   error
	monitorSettingsMu     sync.Mutex
	systemBaseCache       psutil.HostInfo
	resourceCache         psutil.Snapshot
	disksInfoCache        []psutil.DiskInfo
	netRateCache          map[string]psutil.NetworkIO
	diskRateCache         map[string]psutil.DiskIO
	monitorSequence       int64
	monitorUpdatedAt      int64
	monitorHistory        []monitorSample
	monitorArchive        *datastore.Timeline
	monitorRecorder       *datastore.Recorder
	monitorInitError      error
	monitorDisks          map[string]bool
	monitorNetworks       map[string]bool
	monitorDiskReady      bool
}

// NewService 实例化service
func NewService(configService serviceConfig.Service, daemon *daemon.Daemon, container *container.Manager) *service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		ctx:                    ctx,
		cancel:                 cancel,
		panelStartedAt:         time.Now(),
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

// Close 取消排队与执行中的任务，等待监控和后台协程退出后保存历史，可重复调用。
func (s *service) Close() {
	s.closeOnce.Do(func() {
		s.lifecycleMu.Lock()
		s.closed = true
		s.cancel()
		s.lifecycleMu.Unlock()

		s.mu.RLock()
		ids := make([]string, 0, len(s.tasks))
		for id, task := range s.tasks {
			if task.Status == system_task_status.Pending || task.Status == system_task_status.Running {
				ids = append(ids, id)
			}
		}
		s.mu.RUnlock()
		for _, id := range ids {
			s.TaskCancel(id)
		}

		s.queueMu.Lock()
		clear(s.jobs)
		s.queue = nil
		s.queueCond.Broadcast()
		s.queueMu.Unlock()
		s.wait.Wait()
		// 主记录器排空后再关闭进程历史，保留最后一批成功指标的进程记录。
		if s.monitorRecorder != nil {
			_ = s.monitorRecorder.Close()
		}
		if s.monitorProcessStore != nil {
			_ = s.monitorProcessStore.Close()
		}
		if s.monitorArchive != nil {
			_ = s.monitorArchive.Close()
		}
	})
}
