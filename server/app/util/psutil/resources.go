package psutil

import (
	"context"
	"errors"
	"runtime"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
)

// Host 获取主机与系统信息；失败时仍返回当前编译平台等已知信息。
func Host(ctx context.Context) (HostInfo, error) {
	result := HostInfo{
		Hostname:        "Unknown",
		OS:              runtime.GOOS,
		Platform:        "Unknown",
		PlatformVersion: "Unknown",
		KernelVersion:   "Unknown",
		KernelArch:      runtime.GOARCH,
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	value, err := host.InfoWithContext(ctx)
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return result, err
	}
	result.Hostname = value.Hostname
	result.Platform = value.Platform
	result.PlatformVersion = value.PlatformVersion
	result.KernelVersion = value.KernelVersion
	result.KernelArch = value.KernelArch
	result.Uptime = value.Uptime
	result.BootTime = value.BootTime
	return result, nil
}

// CPU 获取 CPU 型号、核心数量和累计 CPU 时间；实时使用率通过 Sampler.Sample 获取。
func CPU(ctx context.Context) (CPUInfo, error) {
	result := CPUInfo{
		Model:     "Unknown",
		Processor: runtime.NumCPU(),
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	info, infoErr := cpu.InfoWithContext(ctx)
	if len(info) > 0 {
		result.Model = info[0].ModelName
	}
	var countErr error
	result.Cores, countErr = cpu.CountsWithContext(ctx, true)
	total, busy, timesErr := cpuTimes(ctx)
	result.TotalTime, result.BusyTime = total, busy
	return result, errors.Join(infoErr, countErr, timesErr, ctx.Err())
}

// cpuTimes 避免读取 gopsutil 包级 CPU 百分比基线，不同 Sampler 之间不会相互影响。
func cpuTimes(ctx context.Context) (float64, float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}
	values, err := cpu.TimesWithContext(ctx, false)
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return 0, 0, err
	}
	if len(values) == 0 {
		return 0, 0, errors.New("未读取到 CPU 计数器")
	}
	value := values[0]
	// Guest 已包含在 User/Nice 内，不再重复计入。
	total := value.User + value.System + value.Nice + value.Iowait + value.Irq + value.Softirq + value.Steal + value.Idle
	return total, total - value.Idle - value.Iowait, nil
}

// Memory 获取物理内存信息。
func Memory(ctx context.Context) (MemoryInfo, error) {
	if err := ctx.Err(); err != nil {
		return MemoryInfo{}, err
	}
	value, err := mem.VirtualMemoryWithContext(ctx)
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return MemoryInfo{}, err
	}
	return MemoryInfo{
		Available:      true,
		Total:          value.Total,
		Used:           value.Used,
		Free:           value.Free,
		AvailableBytes: value.Available,
		UsedPercent:    value.UsedPercent,
	}, nil
}

// Load 获取系统原生平均负载。Windows 不提供 Unix 负载，返回错误；Sampler 提供估算值。
func Load(ctx context.Context) (LoadInfo, error) {
	if err := ctx.Err(); err != nil {
		return LoadInfo{}, err
	}
	value, err := load.AvgWithContext(ctx)
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return LoadInfo{}, err
	}
	return LoadInfo{
		Available: true,
		Load1:     value.Load1,
		Load5:     value.Load5,
		Load15:    value.Load15,
	}, nil
}

// DiskCounters 返回磁盘累计 I/O，按设备名排序；不计算速率、不缓存。
func DiskCounters(ctx context.Context) ([]DiskIO, error) {
	if err := ctx.Err(); err != nil {
		return []DiskIO{}, err
	}
	values, err := disk.IOCountersWithContext(ctx)
	err = errors.Join(err, ctx.Err())
	result := make([]DiskIO, 0, len(values))
	if err != nil {
		return result, err
	}
	for name, value := range values {
		latency := 0.0
		if operations := value.ReadCount + value.WriteCount; operations > 0 {
			latency = float64(value.ReadTime+value.WriteTime) / float64(operations)
		}
		result = append(result, DiskIO{
			Name:         name,
			SerialNumber: value.SerialNumber,
			ReadCount:    value.ReadCount,
			WriteCount:   value.WriteCount,
			ReadBytes:    value.ReadBytes,
			WriteBytes:   value.WriteBytes,
			ReadTime:     value.ReadTime,
			WriteTime:    value.WriteTime,
			IOTime:       value.IoTime,
			WeightedIO:   value.WeightedIO,
			IOLatency:    latency,
		})
	}
	slices.SortFunc(result, func(a, b DiskIO) int { return strings.Compare(a.Name, b.Name) })
	if len(result) == 0 {
		return result, errors.New("未读取到磁盘计数器")
	}
	return result, nil
}

// NetworkCounters 返回所有网卡累计流量和状态，按名称排序；不计算速率、不缓存。
func NetworkCounters(ctx context.Context) ([]NetworkIO, error) {
	if err := ctx.Err(); err != nil {
		return []NetworkIO{}, err
	}
	values, err := gnet.IOCountersWithContext(ctx, true)
	err = errors.Join(err, ctx.Err())
	result := make([]NetworkIO, 0, len(values))
	if err != nil {
		return result, err
	}
	interfaces, interfaceErr := gnet.InterfacesWithContext(ctx)
	flags := make(map[string][]string, len(interfaces))
	for _, item := range interfaces {
		flags[item.Name] = item.Flags
	}
	for _, value := range values {
		name := strings.ToLower(value.Name)
		_, stateKnown := flags[value.Name]
		result = append(result, NetworkIO{
			Name:           value.Name,
			BytesSent:      value.BytesSent,
			BytesRecv:      value.BytesRecv,
			PacketsSent:    value.PacketsSent,
			PacketsRecv:    value.PacketsRecv,
			Errin:          value.Errin,
			Errout:         value.Errout,
			Dropin:         value.Dropin,
			Dropout:        value.Dropout,
			StateAvailable: interfaceErr == nil && stateKnown,
			IsUp:           slices.Contains(flags[value.Name], "up"),
			IsLoopback:     slices.Contains(flags[value.Name], "loopback") || name == "lo" || name == "lo0" || strings.Contains(name, "loopback"),
		})
	}
	slices.SortFunc(result, func(a, b NetworkIO) int { return strings.Compare(a.Name, b.Name) })
	if len(result) == 0 {
		return result, errors.New("未读取到网卡计数器")
	}
	return result, ctx.Err()
}
