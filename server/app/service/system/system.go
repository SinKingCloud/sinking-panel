package system

import (
	"context"
	serviceConfig "server/app/service/config"
	serviceFile "server/app/service/file"
	"server/app/util/cache"
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
	Start()
	Close()
	Restart(action string, containers bool) error                            // 重启面板或服务器，可选同时重启容器
	GetInfo() map[string]interface{}                                         // 主机、系统和运行时信息
	GetWebInfo() map[string]interface{}                                      // 网站名称和界面设置
	GetTimezone() (map[string]interface{}, error)                            // 本机时区、时区列表和系统当地时间
	SetTimezone(name string) error                                           // 设置系统时区
	SyncTimezone(ctx context.Context) error                                  // 同步服务器时间
	GetStatis() map[string]interface{}                                       // CPU、内存、磁盘等资源统计
	GetIO(netInterface, diskName string, after int64) map[string]interface{} // 网络与磁盘 I/O 统计
	GetMonitorSettings() (datastore.SeriesSettings, error)
	SetMonitorSettings(enabled bool, days int) error
	GetMonitorHistory(ctx context.Context, start, end int64, maxPoints int) (MonitorHistory, error)
}

// service 注入结构
type service struct {
	ctx            context.Context
	cancel         context.CancelFunc
	wait           sync.WaitGroup
	closeOnce      sync.Once
	lifecycleMu    sync.Mutex
	started        bool
	closed         bool
	panelStartedAt time.Time
	configService  serviceConfig.Service
	fileService    serviceFile.Service
	cache          cache.Interface
	logoMu         sync.RWMutex
	daemon         *daemon.Daemon
	container      *container.Manager
	restartPending atomic.Bool

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
func NewService(configService serviceConfig.Service, fileService serviceFile.Service, daemon *daemon.Daemon, container *container.Manager, cache cache.Interface) *service {
	ctx, cancel := context.WithCancel(context.Background())
	return &service{
		ctx:             ctx,
		cancel:          cancel,
		panelStartedAt:  time.Now(),
		configService:   configService,
		fileService:     fileService,
		cache:           cache,
		daemon:          daemon,
		container:       container,
		resourceSampler: psutil.NewSampler(),
		disksInfoCache:  make([]psutil.DiskInfo, 0),
		resourceCache:   psutil.Snapshot{Network: make([]psutil.NetworkIO, 0)},
		netRateCache:    make(map[string]psutil.NetworkIO),
		diskRateCache:   make(map[string]psutil.DiskIO),
		monitorHistory:  make([]monitorSample, 0, 120),
	}
}

// Start 启动系统监控。
func (s *service) Start() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.started || s.closed {
		return
	}
	s.started = true
	s.startMonitor()
}

// Close 停止系统监控和后台协程，保存监控历史，可重复调用。
func (s *service) Close() {
	s.closeOnce.Do(func() {
		s.lifecycleMu.Lock()
		s.closed = true
		s.cancel()
		s.lifecycleMu.Unlock()

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
