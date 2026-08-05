package system

import (
	"runtime"
	"server/app/constant"
	"server/app/service/file"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// startMonitor 启动系统监控
func (s *service) startMonitor() {
	// 首次同步更新，确保服务启动后即可读取监控数据
	s.updateMonitor(true)

	// 只保留一个后台goroutine定期更新所有监控数据
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		count := 0
		for range ticker.C {
			count++
			s.updateMonitor(count%5 == 0)
		}
	}()
}

// updateMonitor 更新全部系统监控信息
func (s *service) updateMonitor(refreshStatic bool) {
	now := time.Now()
	interval := now.Sub(s.lastUpdateTime).Seconds()
	if interval <= 0 {
		interval = 1
	}

	// 网卡统计只采集一次，同时用于网卡信息与速率计算
	netCounters, netErr := net.IOCounters(true)
	diskCounters, diskErr := disk.IOCounters()
	cpuInfo := s.getCpuInfo()
	cpuUsage, _ := cpuInfo["usage"].(float64)

	memoryInfo := s.getMemoryInfo()
	loadInfo := s.getLoadInfo(cpuUsage)
	runtimeInfo := s.getRuntimeInfo()
	networkInfo := s.getNetworkInfo(netCounters)

	s.statusCacheLock.Lock()
	hasPreviousCounter := len(s.netIOCache) > 0 || len(s.diskIOCache) > 0
	updated := false
	if netErr == nil {
		s.updateNetworkRate(netCounters, interval)
		netIOCache := make(map[string]net.IOCountersStat, len(netCounters))
		for _, counter := range netCounters {
			netIOCache[counter.Name] = counter
		}
		s.netIOCache = netIOCache
		updated = true
	}

	if diskErr == nil {
		s.updateDiskIORate(diskCounters, interval)
		s.diskIOCache = diskCounters
		updated = true
	}

	if updated {
		s.lastUpdateTime = now
		if hasPreviousCounter {
			s.monitorUpdatedAt = now.UnixMilli()
		}
	}
	s.statusCacheLock.Unlock()

	s.cpuInfoCacheLock.Lock()
	s.cpuInfoCache = cpuInfo
	s.cpuInfoCacheLock.Unlock()

	s.memoryInfoCacheLock.Lock()
	s.memoryInfoCache = memoryInfo
	s.memoryInfoCacheLock.Unlock()

	s.loadInfoCacheLock.Lock()
	s.loadInfoCache = loadInfo
	s.loadInfoCacheLock.Unlock()

	s.runtimeInfoCacheLock.Lock()
	s.runtimeInfoCache = runtimeInfo
	s.runtimeInfoCacheLock.Unlock()

	s.networkInfoCacheLock.Lock()
	s.networkInfoCache = networkInfo
	s.networkInfoCacheLock.Unlock()

	if refreshStatic {
		systemBase := s.getSystemBaseInfo()
		disksInfo := s.getDisksInfo()

		s.systemBaseCacheLock.Lock()
		s.systemBaseCache = systemBase
		s.systemBaseCacheLock.Unlock()

		s.disksInfoCacheLock.Lock()
		s.disksInfoCache = disksInfo
		s.disksInfoCacheLock.Unlock()
	}
}

// updateNetworkRate 更新网卡速率
func (s *service) updateNetworkRate(netCounters []net.IOCountersStat, interval float64) {
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
		if prevCounter, exists := s.netIOCache[counter.Name]; exists {
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

	s.netRateCache = newNetRateCache
}

// updateDiskIORate 更新磁盘IO速率
func (s *service) updateDiskIORate(diskCounters map[string]disk.IOCountersStat, interval float64) {
	// 新的磁盘IO速率数据
	newDiskRateCache := make(map[string]map[string]interface{})

	// 计算每个磁盘的IO速率
	for name, counter := range diskCounters {
		// 构建磁盘IO状态
		diskStats := s.createBasicDiskIOStats(counter)

		// 如果有上一次数据，计算速率
		if prevCounter, exists := s.diskIOCache[name]; exists {
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

	s.diskRateCache = newDiskRateCache
}

// 以下是基础信息获取方法

// getSystemBaseInfo 获取系统基本信息
func (s *service) getSystemBaseInfo() map[string]interface{} {
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
		"version":          "Unknown",      // 系统版本
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
			"version":          constant.Version,         // 系统版本
		}
	}

	return data
}

// getCpuInfo 获取CPU信息
func (s *service) getCpuInfo() map[string]interface{} {
	if !s.cpuInfoInitialized {
		cpuInfo, _ := cpu.Info()
		s.cpuModel = "Unknown"
		if len(cpuInfo) > 0 {
			s.cpuModel = cpuInfo[0].ModelName
		}
		s.cpuCores, _ = cpu.Counts(true)
		s.cpuInfoInitialized = true
	}

	// 与上次监控周期的CPU时间进行比较，不额外阻塞采样
	cpuPercent, _ := cpu.Percent(0, false)

	// 获取CPU使用率
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	return map[string]interface{}{
		"model":     s.cpuModel,       // CPU型号
		"cores":     s.cpuCores,       // CPU核心数
		"usage":     cpuUsage,         // CPU使用率(百分比)
		"processor": runtime.NumCPU(), // CPU逻辑处理器数量
	}
}

// getMemoryInfo 获取内存信息
func (s *service) getMemoryInfo() map[string]interface{} {
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
func (s *service) getDisksInfo() []file.Disk {
	// 磁盘信息
	disks, _ := s.fileService.GetDisks()
	return disks
}

// getLoadInfo 获取系统负载信息
func (s *service) getLoadInfo(cpuUsage float64) map[string]interface{} {
	// 负载信息默认值
	data := map[string]interface{}{
		"load1":  0.0, // 1分钟平均负载
		"load5":  0.0, // 5分钟平均负载
		"load15": 0.0, // 15分钟平均负载
	}

	// Windows系统不支持原生load.Avg，需要模拟实现
	if runtime.GOOS == "windows" {
		// Windows复用本轮CPU利用率，避免为模拟负载再次阻塞采样
		loadValue := cpuUsage / 100.0 * float64(runtime.NumCPU())
		data = map[string]interface{}{
			"load1":  loadValue,
			"load5":  loadValue,
			"load15": loadValue,
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
func (s *service) getRuntimeInfo() map[string]interface{} {
	return map[string]interface{}{
		"go_version":    runtime.Version(),      // Go语言版本
		"num_goroutine": runtime.NumGoroutine(), // 当前Goroutine数量
	}
}

// getNetworkInfo 获取网卡信息
func (s *service) getNetworkInfo(netStats []net.IOCountersStat) []map[string]interface{} {
	var networkInfo []map[string]interface{}

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
func (s *service) createBasicDiskIOStats(counter disk.IOCountersStat) map[string]interface{} {
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
