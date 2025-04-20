package system

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"runtime"
	"server/app/service/file"
	"sync"
	"time"
)

// 系统监控数据缓存
var (
	// 系统基本信息缓存
	systemBaseCache     map[string]interface{} // 系统基本信息
	systemBaseCacheLock sync.RWMutex

	// CPU信息缓存
	cpuInfoCache     map[string]interface{} // CPU信息
	cpuInfoCacheLock sync.RWMutex

	// 内存信息缓存
	memoryInfoCache     map[string]interface{} // 内存信息
	memoryInfoCacheLock sync.RWMutex

	// 磁盘信息缓存
	disksInfoCache     []file.Disk // 磁盘信息
	disksInfoCacheLock sync.RWMutex

	// 系统负载信息缓存
	loadInfoCache     map[string]interface{} // 系统负载信息
	loadInfoCacheLock sync.RWMutex

	// 运行时信息缓存
	runtimeInfoCache     map[string]interface{} // 运行时信息
	runtimeInfoCacheLock sync.RWMutex

	// 网卡基本信息缓存
	networkInfoCache     []map[string]interface{} // 网卡信息
	networkInfoCacheLock sync.RWMutex

	// 网卡流量缓存
	netIOCache     map[string]net.IOCountersStat
	netIOCacheLock sync.RWMutex

	// 磁盘IO缓存
	diskIOCache     map[string]disk.IOCountersStat
	diskIOCacheLock sync.RWMutex

	// 网卡速率数据缓存
	netRateCache     map[string]map[string]interface{}
	netRateCacheLock sync.RWMutex

	// 磁盘IO速率数据缓存
	diskRateCache     map[string]map[string]interface{}
	diskRateCacheLock sync.RWMutex

	// 上次更新时间
	lastUpdateTime     time.Time
	lastUpdateTimeLock sync.RWMutex
)

// 初始化，在包被导入时自动执行
func init() {
	// 初始化各个缓存
	systemBaseCache = make(map[string]interface{})
	cpuInfoCache = make(map[string]interface{})
	memoryInfoCache = make(map[string]interface{})
	disksInfoCache = make([]file.Disk, 0)
	loadInfoCache = make(map[string]interface{})
	runtimeInfoCache = make(map[string]interface{})
	networkInfoCache = make([]map[string]interface{}, 0)
	netIOCache = make(map[string]net.IOCountersStat)
	diskIOCache = make(map[string]disk.IOCountersStat)
	netRateCache = make(map[string]map[string]interface{})
	diskRateCache = make(map[string]map[string]interface{})
	lastUpdateTime = time.Now()

	// 获取单例对象
	s := GetIns()

	// 首次更新系统数据
	s.updateSystemMonitor()
	s.updateIOMonitor()

	// 启动后台goroutine定期更新系统数据
	go func() {
		// 统一使用3秒的更新频率
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// 监控更新循环
		for {
			select {
			case <-ticker.C:
				// 更新所有系统数据
				s.updateSystemMonitor()
				s.updateIOMonitor()
			}
		}
	}()
}

// updateSystemMonitor 更新系统基础监控信息
func (s *Service) updateSystemMonitor() {
	// 并发获取各类系统信息
	var wg sync.WaitGroup
	wg.Add(7)

	// 系统基本信息
	go func() {
		defer wg.Done()
		systemBase := s.getSystemBaseInfo()

		systemBaseCacheLock.Lock()
		systemBaseCache = systemBase
		systemBaseCacheLock.Unlock()
	}()

	// CPU信息
	go func() {
		defer wg.Done()
		cpuInfo := s.getCpuInfo()

		cpuInfoCacheLock.Lock()
		cpuInfoCache = cpuInfo
		cpuInfoCacheLock.Unlock()
	}()

	// 内存信息
	go func() {
		defer wg.Done()
		memoryInfo := s.getMemoryInfo()

		memoryInfoCacheLock.Lock()
		memoryInfoCache = memoryInfo
		memoryInfoCacheLock.Unlock()
	}()

	// 磁盘信息
	go func() {
		defer wg.Done()
		disksInfo := s.getDisksInfo()

		disksInfoCacheLock.Lock()
		disksInfoCache = disksInfo
		disksInfoCacheLock.Unlock()
	}()

	// 系统负载信息
	go func() {
		defer wg.Done()
		loadInfo := s.getLoadInfo()

		loadInfoCacheLock.Lock()
		loadInfoCache = loadInfo
		loadInfoCacheLock.Unlock()
	}()

	// 运行时信息
	go func() {
		defer wg.Done()
		runtimeInfo := s.getRuntimeInfo()

		runtimeInfoCacheLock.Lock()
		runtimeInfoCache = runtimeInfo
		runtimeInfoCacheLock.Unlock()
	}()

	// 网卡基本信息
	go func() {
		defer wg.Done()
		networkInfo := s.getNetworkInfo()

		networkInfoCacheLock.Lock()
		networkInfoCache = networkInfo
		networkInfoCacheLock.Unlock()
	}()

	// 等待所有数据收集完成
	wg.Wait()
}

// updateIOMonitor 更新IO监控信息
func (s *Service) updateIOMonitor() {
	// 获取当前时间
	now := time.Now()

	// 获取上次更新时间
	lastUpdateTimeLock.RLock()
	lastTime := lastUpdateTime
	lastUpdateTimeLock.RUnlock()

	// 计算时间间隔（秒）
	interval := now.Sub(lastTime).Seconds()
	if interval <= 0 {
		interval = 1 // 防止除以零
	}

	// 更新网卡和磁盘IO状态
	var wg sync.WaitGroup
	wg.Add(2)

	// 更新网卡IO
	go func() {
		defer wg.Done()
		if netCounters, err := net.IOCounters(true); err == nil {
			// 更新网卡速率
			s.updateNetworkRate(netCounters, interval)

			// 更新网卡数据缓存
			netIOCacheLock.Lock()
			for _, counter := range netCounters {
				netIOCache[counter.Name] = counter
			}
			netIOCacheLock.Unlock()
		}
	}()

	// 更新磁盘IO
	go func() {
		defer wg.Done()
		if diskCounters, err := disk.IOCounters(); err == nil {
			// 更新磁盘IO速率
			s.updateDiskIORate(diskCounters, interval)

			// 更新磁盘IO数据缓存
			diskIOCacheLock.Lock()
			for name, counter := range diskCounters {
				diskIOCache[name] = counter
			}
			diskIOCacheLock.Unlock()
		}
	}()

	// 等待IO数据更新完成
	wg.Wait()

	// 更新时间
	lastUpdateTimeLock.Lock()
	lastUpdateTime = now
	lastUpdateTimeLock.Unlock()
}

// updateNetworkRate 更新网卡速率
func (s *Service) updateNetworkRate(netCounters []net.IOCountersStat, interval float64) {
	// 获取之前的网卡数据
	prevNetIOData := make(map[string]net.IOCountersStat)
	netIOCacheLock.RLock()
	for name, counter := range netIOCache {
		prevNetIOData[name] = counter
	}
	netIOCacheLock.RUnlock()

	// 新的网卡速率数据
	newNetRateCache := make(map[string]map[string]interface{})

	// 计算每个网卡的速率
	for _, counter := range netCounters {
		// 过滤掉一些虚拟网卡（可根据需要调整）
		if (runtime.GOOS == "windows" && (counter.Name == "Loopback Pseudo-Interface 1" ||
			counter.Name == "Loopback" || counter.Name == "lo" || counter.Name == "lo0")) ||
			(runtime.GOOS == "linux" && (counter.Name == "lo" || counter.Name == "lo0")) ||
			(runtime.GOOS == "darwin" && (counter.Name == "lo0")) {
			continue
		}

		// 获取上一次的数据
		var sendRateBps, recvRateBps float64
		if prevCounter, exists := prevNetIOData[counter.Name]; exists {
			// 计算速率（字节/秒）
			sendRateBps = float64(counter.BytesSent-prevCounter.BytesSent) / interval
			recvRateBps = float64(counter.BytesRecv-prevCounter.BytesRecv) / interval
		}

		// 构建网卡速率数据
		newNetRateCache[counter.Name] = map[string]interface{}{
			"bytes_sent":   counter.BytesSent,   // 总发送字节
			"bytes_recv":   counter.BytesRecv,   // 总接收字节
			"packets_sent": counter.PacketsSent, // 总发送数据包
			"packets_recv": counter.PacketsRecv, // 总接收数据包
			"errin":        counter.Errin,       // 接收错误
			"errout":       counter.Errout,      // 发送错误
			"dropin":       counter.Dropin,      // 接收丢包
			"dropout":      counter.Dropout,     // 发送丢包
			"send_rate":    sendRateBps,         // 发送速率 (字节/秒)
			"recv_rate":    recvRateBps,         // 接收速率 (字节/秒)
		}
	}

	// 更新网卡速率缓存
	netRateCacheLock.Lock()
	netRateCache = newNetRateCache
	netRateCacheLock.Unlock()
}

// updateDiskIORate 更新磁盘IO速率
func (s *Service) updateDiskIORate(diskCounters map[string]disk.IOCountersStat, interval float64) {
	// 获取之前的磁盘IO数据
	prevDiskIOData := make(map[string]disk.IOCountersStat)
	diskIOCacheLock.RLock()
	for name, counter := range diskIOCache {
		prevDiskIOData[name] = counter
	}
	diskIOCacheLock.RUnlock()

	// 新的磁盘IO速率数据
	newDiskRateCache := make(map[string]map[string]interface{})

	// 计算每个磁盘的IO速率
	for name, counter := range diskCounters {
		// 构建磁盘IO状态
		diskStats := s.createBasicDiskIOStats(counter)

		// 如果有上一次数据，计算速率
		if prevCounter, exists := prevDiskIOData[name]; exists {
			// 计算速率
			readBytesRate := float64(counter.ReadBytes-prevCounter.ReadBytes) / interval
			writeBytesRate := float64(counter.WriteBytes-prevCounter.WriteBytes) / interval
			readCountRate := float64(counter.ReadCount-prevCounter.ReadCount) / interval
			writeCountRate := float64(counter.WriteCount-prevCounter.WriteCount) / interval

			// 更新速率数据
			diskStats["read_bytes_rate"] = readBytesRate   // 读取速率(字节/秒)
			diskStats["write_bytes_rate"] = writeBytesRate // 写入速率(字节/秒)
			diskStats["read_count_rate"] = readCountRate   // 读取次数速率(次/秒)
			diskStats["write_count_rate"] = writeCountRate // 写入次数速率(次/秒)
		}

		// 添加到磁盘IO速率缓存
		newDiskRateCache[name] = diskStats
	}

	// 更新磁盘IO速率缓存
	diskRateCacheLock.Lock()
	diskRateCache = newDiskRateCache
	diskRateCacheLock.Unlock()
}

// 以下是基础信息获取方法

// getSystemBaseInfo 获取系统基本信息
func (s *Service) getSystemBaseInfo() map[string]interface{} {
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

// getCpuInfo 获取CPU信息
func (s *Service) getCpuInfo() map[string]interface{} {
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

	// 获取CPU型号
	cpuModel := "Unknown"
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	// 获取CPU使用率
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	return map[string]interface{}{
		"model":     cpuModel,         // CPU型号
		"cores":     cpuCounts,        // CPU物理核心数
		"usage":     cpuUsage,         // CPU使用率(百分比)
		"processor": runtime.NumCPU(), // CPU逻辑处理器数量
	}
}

// getMemoryInfo 获取内存信息
func (s *Service) getMemoryInfo() map[string]interface{} {
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

// getDisksInfo 获取磁盘信息
func (s *Service) getDisksInfo() []file.Disk {
	// 磁盘信息
	disks, _ := file.GetIns().GetDisks()
	return disks
}

// getLoadInfo 获取系统负载信息
func (s *Service) getLoadInfo() map[string]interface{} {
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

// getRuntimeInfo 获取运行时信息
func (s *Service) getRuntimeInfo() map[string]interface{} {
	return map[string]interface{}{
		"go_version":    runtime.Version(),      // Go语言版本
		"num_goroutine": runtime.NumGoroutine(), // 当前Goroutine数量
	}
}

// getNetworkInfo 获取网卡信息
func (s *Service) getNetworkInfo() []map[string]interface{} {
	var networkInfo []map[string]interface{}

	// 获取所有网卡统计信息（含流量信息）
	netStats, err := net.IOCounters(true)
	if err != nil {
		return networkInfo
	}

	// 处理网卡流量统计信息
	for _, stat := range netStats {
		name := stat.Name

		// 过滤掉环回接口
		if (runtime.GOOS == "windows" && (name == "Loopback Pseudo-Interface 1" ||
			name == "Loopback" || name == "lo" || name == "lo0")) ||
			(runtime.GOOS == "linux" && (name == "lo" || name == "lo0")) ||
			(runtime.GOOS == "darwin" && (name == "lo0")) {
			continue
		}

		// 创建网卡基本信息
		netInfo := map[string]interface{}{
			"name":         name,             // 网卡名称
			"bytes_sent":   stat.BytesSent,   // 发送的字节数
			"bytes_recv":   stat.BytesRecv,   // 接收的字节数
			"packets_sent": stat.PacketsSent, // 发送的数据包数
			"packets_recv": stat.PacketsRecv, // 接收的数据包数
			"errin":        stat.Errin,       // 接收错误
			"errout":       stat.Errout,      // 发送错误
			"dropin":       stat.Dropin,      // 接收丢包
			"dropout":      stat.Dropout,     // 发送丢包
			"is_up":        true,             // 默认为启用状态
			"is_loopback":  false,            // 已过滤掉环回接口
		}
		// 添加到结果集
		networkInfo = append(networkInfo, netInfo)
	}

	return networkInfo
}

// createBasicDiskIOStats 创建基本磁盘IO统计信息
func (s *Service) createBasicDiskIOStats(counter disk.IOCountersStat) map[string]interface{} {
	// IO延迟 = (读取时间 + 写入时间) / (读取计数 + 写入计数)
	var ioLatency float64
	totalOps := counter.ReadCount + counter.WriteCount
	if totalOps > 0 {
		totalTime := counter.ReadTime + counter.WriteTime
		ioLatency = float64(totalTime) / float64(totalOps)
	}

	return map[string]interface{}{
		"read_count":       counter.ReadCount,    // 读取次数
		"write_count":      counter.WriteCount,   // 写入次数
		"read_bytes":       counter.ReadBytes,    // 读取字节数
		"write_bytes":      counter.WriteBytes,   // 写入字节数
		"read_time":        counter.ReadTime,     // 读取时间
		"write_time":       counter.WriteTime,    // 写入时间
		"io_time":          counter.IoTime,       // IO时间
		"weighted_io":      counter.WeightedIO,   // 加权IO
		"name":             counter.Name,         // 磁盘名称
		"serial_number":    counter.SerialNumber, // 序列号
		"read_bytes_rate":  0,                    // 读取速率(字节/秒)
		"write_bytes_rate": 0,                    // 写入速率(字节/秒)
		"read_count_rate":  0,                    // 读取次数速率(次/秒)
		"write_count_rate": 0,                    // 写入次数速率(次/秒)
		"io_latency":       ioLatency,            // IO延迟(毫秒)
	}
}
