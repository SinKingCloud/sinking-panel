package psutil

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"
)

// Sampler 保存本调用方的差分基线；不创建协程、不等待采样间隔、不写文件。
// 每个监控调用方独立创建一个 Sampler；同一实例的并发调用会串行采样。
type Sampler struct {
	mu             sync.Mutex
	cpuInfo        CPUInfo
	cpuInitialized bool
	cpuPrevious    CPUInfo
	cpuReady       bool
	network        map[string]NetworkIO
	disk           map[string]DiskIO
	networkAt      time.Time
	diskAt         time.Time
	load           [3]float64
	loadAt         time.Time
}

// NewSampler 创建独立资源采样器，首次 Sample 只建立速率和 CPU 使用率基线。
func NewSampler() *Sampler { return &Sampler{} }

// Sample 同步采集 CPU、内存、负载、网卡和磁盘 I/O；调用周期由业务方控制。
// 新设备或计数器回退时重新建立该设备的基线，不产生负速率。
func (s *Sampler) Sample(ctx context.Context) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := Snapshot{
		Time:   time.Now().UnixMilli(),
		Errors: make(map[string]string),
	}
	if !s.cpuInitialized {
		var err error
		s.cpuInfo, err = CPU(ctx)
		s.cpuInitialized = err == nil
		if err != nil {
			result.Errors["cpu"] = err.Error()
		}
	}
	result.CPU = s.cpuInfo
	total, busy, err := cpuTimes(ctx)
	if err != nil {
		result.Errors["cpu"] = err.Error()
	} else {
		result.CPU.TotalTime, result.CPU.BusyTime = total, busy
		previous := s.cpuPrevious
		if s.cpuReady && total > previous.TotalTime && busy >= previous.BusyTime {
			result.CPU.Usage = min(100, max(0, (busy-previous.BusyTime)/(total-previous.TotalTime)*100))
			result.CPU.Available = true
		}
		s.cpuPrevious, s.cpuReady = result.CPU, true
	}
	result.Memory, err = Memory(ctx)
	if err != nil {
		result.Errors["memory"] = err.Error()
	}
	if runtime.GOOS == "windows" {
		result.Load.Estimated = true
		if result.CPU.Available {
			now := time.Now()
			value := result.CPU.Usage / 100 * float64(runtime.NumCPU())
			if s.loadAt.IsZero() {
				s.load = [3]float64{value, value, value}
			} else {
				elapsed := now.Sub(s.loadAt).Seconds()
				for index, window := range [3]float64{60, 300, 900} {
					s.load[index] += (value - s.load[index]) * (1 - math.Exp(-math.Max(0, elapsed)/window))
				}
			}
			s.loadAt = now
			result.Load.Available = true
			result.Load.Load1, result.Load.Load5, result.Load.Load15 = s.load[0], s.load[1], s.load[2]
		}
	} else {
		result.Load, err = Load(ctx)
		if err != nil {
			result.Errors["load"] = err.Error()
		}
	}
	result.Network, err = NetworkCounters(ctx)
	if err != nil {
		result.Errors["network"] = err.Error()
	} else {
		now := time.Now()
		elapsed := now.Sub(s.networkAt).Seconds()
		next := make(map[string]NetworkIO, len(result.Network))
		for index := range result.Network {
			value := &result.Network[index]
			if previous, ok := s.network[value.Name]; ok && elapsed > 0 && value.BytesSent >= previous.BytesSent && value.BytesRecv >= previous.BytesRecv {
				value.SendRate = float64(value.BytesSent-previous.BytesSent) / elapsed
				value.RecvRate = float64(value.BytesRecv-previous.BytesRecv) / elapsed
				value.RateAvailable = true
			}
			next[value.Name] = *value
		}
		s.network, s.networkAt = next, now
	}
	result.Disk, err = DiskCounters(ctx)
	if err != nil {
		result.Errors["disk"] = err.Error()
	} else {
		now := time.Now()
		elapsed := now.Sub(s.diskAt).Seconds()
		next := make(map[string]DiskIO, len(result.Disk))
		for index := range result.Disk {
			value := &result.Disk[index]
			if previous, ok := s.disk[value.Name]; ok && elapsed > 0 && value.ReadBytes >= previous.ReadBytes && value.WriteBytes >= previous.WriteBytes && value.ReadCount >= previous.ReadCount && value.WriteCount >= previous.WriteCount {
				value.ReadBytesRate = float64(value.ReadBytes-previous.ReadBytes) / elapsed
				value.WriteBytesRate = float64(value.WriteBytes-previous.WriteBytes) / elapsed
				value.ReadCountRate = float64(value.ReadCount-previous.ReadCount) / elapsed
				value.WriteCountRate = float64(value.WriteCount-previous.WriteCount) / elapsed
				value.RateAvailable = true
			}
			next[value.Name] = *value
		}
		s.disk, s.diskAt = next, now
	}
	return result
}
