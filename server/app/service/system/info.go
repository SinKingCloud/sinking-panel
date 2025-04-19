package system

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"runtime"
	"server/app/service/file"
	"sync"
	"time"
)

// 系统信息缓存
var (
	systemInfoCache     map[string]interface{} // 缓存的系统信息
	systemInfoCacheLock sync.RWMutex           // 读写锁
)

// 初始化，在包被导入时自动执行
func init() {
	// 初始化缓存
	systemInfoCache = make(map[string]interface{})

	// 获取单例对象
	s := GetIns()

	// 首次获取系统信息
	s.updateSystemInfo()

	// 启动后台goroutine定期更新系统信息
	go func() {
		ticker := time.NewTicker(5 * time.Second) // 每5秒更新一次
		defer ticker.Stop()

		for range ticker.C {
			s.updateSystemInfo()
		}
	}()
}

// updateSystemInfo 更新系统信息
func (s *Service) updateSystemInfo() {
	// 构建数据
	data := map[string]interface{}{
		"system":  s.GetSystemBaseInfo(), // 系统基本信息
		"cpu":     s.GetCpuInfo(),        // CPU信息
		"memory":  s.GetMemoryInfo(),     // 内存信息
		"disks":   s.GetDisksInfo(),      // 磁盘信息
		"load":    s.GetLoadInfo(),       // 系统负载信息
		"runtime": s.GetRuntimeInfo(),    // 运行时信息
	}
	// 更新缓存
	systemInfoCacheLock.Lock()
	systemInfoCache = data
	systemInfoCacheLock.Unlock()
}

// GetSystemBaseInfo 获取系统基本信息
func (s *Service) GetSystemBaseInfo() map[string]interface{} {
	// 系统信息默认值
	data := map[string]interface{}{
		"hostname":         "Unknown",      // 主机名
		"os":               runtime.GOOS,   // 操作系统类型
		"platform":         "Unknown",      // 平台名称(如Windows、Linux等)
		"platform_version": "Unknown",      // 平台版本(如Windows 10、CentOS 7等)
		"kernel_version":   "Unknown",      // 内核版本
		"kernel_arch":      runtime.GOARCH, // 内核架构(如x86_64、arm64等)
		"uptime":           uint64(0),      // 系统运行时间(单位:秒)
		"boot_time":        uint64(0),      // 系统启动时间(时间戳)
	}

	// 获取系统信息（错误处理）
	if hostInfo, err := host.Info(); err == nil {
		data = map[string]interface{}{
			"hostname":         hostInfo.Hostname,        // 主机名
			"os":               runtime.GOOS,             // 操作系统类型
			"platform":         hostInfo.Platform,        // 平台名称(如Windows、Linux等)
			"platform_version": hostInfo.PlatformVersion, // 平台版本(如Windows 10、CentOS 7等)
			"kernel_version":   hostInfo.KernelVersion,   // 内核版本
			"kernel_arch":      hostInfo.KernelArch,      // 内核架构(如x86_64、arm64等)
			"uptime":           hostInfo.Uptime,          // 系统运行时间(单位:秒)
			"boot_time":        hostInfo.BootTime,        // 系统启动时间(时间戳)
		}
	}

	return data
}

// GetCpuInfo 获取CPU信息
func (s *Service) GetCpuInfo() map[string]interface{} {
	var cpuInfo []cpu.InfoStat
	var cpuCounts int
	var cpuPercent []float64

	// 并发获取CPU信息
	var wg sync.WaitGroup
	wg.Add(3)

	// 获取CPU基本信息
	go func() {
		defer wg.Done()
		cpuInfo, _ = cpu.Info()
	}()

	// 获取CPU核心数
	go func() {
		defer wg.Done()
		cpuCounts, _ = cpu.Counts(true)
	}()

	// 获取CPU使用率（使用10毫秒时间窗口代替1秒）
	go func() {
		defer wg.Done()
		cpuPercent, _ = cpu.Percent(10*time.Millisecond, false)
	}()

	// 等待CPU信息获取完成
	wg.Wait()

	return map[string]interface{}{
		"model":     s.getCpuModel(cpuInfo),    // CPU型号
		"cores":     cpuCounts,                 // CPU物理核心数
		"usage":     s.getCpuUsage(cpuPercent), // CPU使用率(百分比)
		"processor": runtime.NumCPU(),          // CPU逻辑处理器数量
	}
}

// GetMemoryInfo 获取内存信息
func (s *Service) GetMemoryInfo() map[string]interface{} {
	// 内存信息默认值
	data := map[string]interface{}{
		"total":        uint64(0), // 总内存(字节)
		"used":         uint64(0), // 已使用内存(字节)
		"free":         uint64(0), // 空闲内存(字节)
		"used_percent": 0.0,       // 内存使用率(百分比)
	}

	// 获取内存信息（错误处理）
	if memInfo, err := mem.VirtualMemory(); err == nil {
		data = map[string]interface{}{
			"total":        memInfo.Total,       // 总内存(字节)
			"used":         memInfo.Used,        // 已使用内存(字节)
			"free":         memInfo.Free,        // 空闲内存(字节)
			"used_percent": memInfo.UsedPercent, // 内存使用率(百分比)
		}
	}

	return data
}

// GetDisksInfo 获取磁盘信息
func (s *Service) GetDisksInfo() []file.Disk {
	// 磁盘信息
	disks, _ := file.GetIns().GetDisks()
	return disks
}

// GetLoadInfo 获取系统负载信息
func (s *Service) GetLoadInfo() map[string]interface{} {
	// 负载信息默认值
	data := map[string]interface{}{
		"load1":  0.0, // 1分钟平均负载
		"load5":  0.0, // 5分钟平均负载
		"load15": 0.0, // 15分钟平均负载
	}

	// Windows系统不支持原生load.Avg，需要模拟实现
	if runtime.GOOS == "windows" {
		// 在Windows上使用CPU利用率作为负载的替代指标
		// 获取当前CPU使用率
		cpuPercent, err := cpu.Percent(1*time.Second, false)
		if err == nil && len(cpuPercent) > 0 {
			// 将CPU利用率转换为类似负载值
			// 在Windows中，可以将CPU利用率除以核心数来近似负载
			// 实际负载 ≈ CPU利用率 / 100 * 逻辑处理器数量
			processors := float64(runtime.NumCPU())
			loadValue := cpuPercent[0] / 100.0 * processors

			// 将计算结果设置为所有时间段的负载值
			// 由于Windows无法获取过去的数据，所以三个值相同
			data = map[string]interface{}{
				"load1":  loadValue, // 使用CPU利用率模拟的负载
				"load5":  loadValue, // 在Windows上无法获取5分钟值，使用相同值
				"load15": loadValue, // 在Windows上无法获取15分钟值，使用相同值
			}
		}
	} else {
		// 在Unix/Linux系统上使用原生负载值
		if loadInfo, err := load.Avg(); err == nil {
			data = map[string]interface{}{
				"load1":  loadInfo.Load1,  // 1分钟平均负载
				"load5":  loadInfo.Load5,  // 5分钟平均负载
				"load15": loadInfo.Load15, // 15分钟平均负载
			}
		}
	}

	return data
}

// GetRuntimeInfo 获取运行时信息
func (s *Service) GetRuntimeInfo() map[string]interface{} {
	return map[string]interface{}{
		"go_version":    runtime.Version(),      // Go语言版本
		"num_goroutine": runtime.NumGoroutine(), // 当前Goroutine数量
	}
}

// GetInfo 获取系统信息
func (s *Service) GetInfo() map[string]interface{} {
	// 从缓存中读取系统信息
	systemInfoCacheLock.RLock()
	data := systemInfoCache
	systemInfoCacheLock.RUnlock()
	return data
}

// getCpuModel 获取CPU型号
func (s *Service) getCpuModel(cpuInfo []cpu.InfoStat) string {
	if len(cpuInfo) > 0 {
		return cpuInfo[0].ModelName
	}
	return "Unknown"
}

// getCpuUsage 获取CPU使用率
func (s *Service) getCpuUsage(cpuPercent []float64) float64 {
	if len(cpuPercent) > 0 {
		return cpuPercent[0]
	}
	return 0.0
}
