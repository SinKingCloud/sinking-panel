package psutil

import (
	"context"
	"math"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

const (
	processTopLimit      = 5
	processScanLimit     = 10000
	processBaselineLimit = 100000
	processBaselineTTL   = 2 * time.Minute
)

// 进程创建时间会使用系统启动时间；初始化时开启库缓存，避免 Linux 每 PID 重读 /proc/stat。
// 只在包初始化时设置，运行中的采样不修改 gopsutil 的全局开关。
func init() { process.EnableBootTimeCache(true) }

// ProcessTopItem 仅返回名称和用量；CPU 为整机归一化 0–100%，内存为字节，磁盘速率为字节/秒。
type ProcessTopItem struct {
	PID          int32   `json:"pid"`
	CreateTime   int64   `json:"create_time"`
	Name         string  `json:"name"`
	CPUPercent   float64 `json:"cpu"`
	MemoryRSS    uint64  `json:"memory"`
	ReadRate     float64 `json:"disk_read"`
	WriteRate    float64 `json:"disk_write"`
	CPUAvailable bool    `json:"cpu_available"`
	IOAvailable  bool    `json:"io_available"`
}

// ProcessTopSnapshot 每榜最多 5 项，按用量降序、PID 升序；Partial 表示结果不完整。
type ProcessTopSnapshot struct {
	Time            int64            `json:"time"`
	CPU             []ProcessTopItem `json:"cpu"`
	Memory          []ProcessTopItem `json:"memory"`
	IO              []ProcessTopItem `json:"disk_io"`
	CPUAvailable    bool             `json:"cpu_available"`
	MemoryAvailable bool             `json:"memory_available"`
	IOAvailable     bool             `json:"io_available"`
	Partial         bool             `json:"partial"`
	Scanned         int              `json:"scanned"`
	Skipped         int              `json:"skipped"`
	Error           string           `json:"error,omitempty"`
}

// Clone 返回独立榜单切片，防止调用方修改已经发布的缓存。
func (s ProcessTopSnapshot) Clone() ProcessTopSnapshot {
	s.CPU, s.Memory, s.IO = slices.Clone(s.CPU), slices.Clone(s.Memory), slices.Clone(s.IO)
	return s
}

type processCounter struct {
	item          ProcessTopItem
	cpu, totalCPU float64
	read, write   uint64
	at, nameAt    time.Time
	cpuOK, ioOK   bool
}

// ProcessSampler 包装 gopsutil 的同步读取和差分，不安排周期任务或启动后台协程。
// 同一实例串行采样；预算为 Linux 200ms、macOS/Windows 500ms，按 PID 轮转避免遗漏后续进程。
type ProcessSampler struct {
	mu       sync.Mutex
	previous map[int32]processCounter
	cursor   int32
}

func NewProcessSampler() *ProcessSampler { return &ProcessSampler{} }

func (s *ProcessSampler) Sample(ctx context.Context) ProcessTopSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := ProcessTopSnapshot{Time: time.Now().UnixMilli(), CPU: make([]ProcessTopItem, 0, processTopLimit), Memory: make([]ProcessTopItem, 0, processTopLimit), IO: make([]ProcessTopItem, 0, processTopLimit)}
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		result.Partial, result.Error = true, "当前平台暂不支持进程排行榜采集"
		return result
	}
	budget := 500 * time.Millisecond
	if runtime.GOOS == "linux" {
		budget = 200 * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	total, _, err := cpuTimes(ctx)
	if err != nil {
		result.Partial, result.Error = true, err.Error()
		return result
	}
	pids, err := process.PidsWithContext(ctx) // 库已按 PID 排序，不重复排序所有进程。
	if err != nil || ctx.Err() != nil {
		if err == nil {
			err = ctx.Err()
		}
		result.Partial, result.Error = true, err.Error()
		return result
	}
	pids = slices.DeleteFunc(pids, func(pid int32) bool { return pid <= 0 })
	if s.previous == nil {
		s.previous = make(map[int32]processCounter)
	}
	now := time.Now()
	for pid, previous := range s.previous {
		_, visible := slices.BinarySearch(pids, pid)
		if !visible || now.Sub(previous.at) > processBaselineTTL {
			delete(s.previous, pid)
		}
	}
	hasBaseline := len(s.previous) > 0
	start, found := slices.BinarySearch(pids, s.cursor)
	if found {
		start++
	}
	for offset := 0; offset < len(pids); offset++ {
		if ctx.Err() != nil || result.Scanned >= processScanLimit {
			result.Partial, result.Error = true, "进程扫描预算或上下文已结束，下轮继续后续 PID"
			break
		}
		pid := pids[(start+offset)%len(pids)]
		s.cursor = pid
		result.Scanned++
		p := &process.Process{Pid: pid} // 已枚举 PID，不重复调用 NewProcess 的存在性探测。
		created, err := p.CreateTimeWithContext(ctx)
		if err != nil || created <= 0 || ctx.Err() != nil {
			delete(s.previous, pid)
			result.Skipped++
			result.Partial = true
			continue
		}
		current := processCounter{item: ProcessTopItem{PID: pid, CreateTime: created}, totalCPU: total, at: time.Now()}
		if times, err := p.TimesWithContext(ctx); err == nil && times != nil {
			current.cpu = times.User + times.System
			current.cpuOK = current.cpu >= 0 && !math.IsNaN(current.cpu) && !math.IsInf(current.cpu, 0)
		}
		memoryOK := false
		if ctx.Err() == nil {
			if memory, err := p.MemoryInfoWithContext(ctx); err == nil && memory != nil {
				current.item.MemoryRSS, memoryOK = memory.RSS, true
			}
		}
		if runtime.GOOS == "darwin" {
			// 此版本库的 taskinfo 权限失败可能返回零结构、nil 错误，不能展示为有效零用量。
			current.cpuOK = current.cpuOK && current.cpu > 0
			memoryOK = memoryOK && current.item.MemoryRSS > 0
		}
		if runtime.GOOS != "windows" && ctx.Err() == nil {
			if counters, err := p.IOCountersWithContext(ctx); err == nil && counters != nil {
				// Linux 的 ReadBytes/WriteBytes 是逻辑 I/O；物理磁盘与 Darwin 均使用 Disk 字段。
				current.read, current.write, current.ioOK = counters.DiskReadBytes, counters.DiskWriteBytes, true
			}
		}
		if ctx.Err() != nil {
			break
		}
		if !current.cpuOK && !memoryOK && !current.ioOK {
			// 没有任何可发布指标时无需再次查身份，也不保留无法参与差分的基线。
			delete(s.previous, pid)
			result.Skipped++
			result.Partial = true
			continue
		}
		// CreateTime 在 Process 对象内有缓存，必须用新对象复核，防止混入复用 PID 的数据。
		check, err := (&process.Process{Pid: pid}).CreateTimeWithContext(ctx)
		if err != nil || check != created || ctx.Err() != nil {
			delete(s.previous, pid)
			result.Skipped++
			result.Partial = true
			continue
		}
		previous, exists := s.previous[pid]
		if exists && previous.item.CreateTime == created {
			current.item.Name, current.nameAt = previous.item.Name, previous.nameAt
			if current.cpuOK && previous.cpuOK && current.cpu >= previous.cpu && total > previous.totalCPU {
				current.item.CPUPercent = min(100, max(0, (current.cpu-previous.cpu)/(total-previous.totalCPU)*100))
				current.item.CPUAvailable = true
			}
			if elapsed := current.at.Sub(previous.at).Seconds(); elapsed > 0 && current.ioOK && previous.ioOK && current.read >= previous.read && current.write >= previous.write {
				current.item.ReadRate = float64(current.read-previous.read) / elapsed
				current.item.WriteRate = float64(current.write-previous.write) / elapsed
				current.item.IOAvailable = true
			}
		}
		if exists || len(s.previous) < processBaselineLimit {
			s.previous[pid] = current
		} else {
			result.Partial = true
		}
		if !current.cpuOK || !memoryOK || !current.ioOK || (hasBaseline && (!current.item.CPUAvailable || !current.item.IOAvailable)) {
			result.Partial = true
		}
		if current.item.CPUAvailable {
			result.CPUAvailable = true
			result.CPU = insertProcessTop(result.CPU, current.item, func(item ProcessTopItem) float64 { return item.CPUPercent })
		}
		if memoryOK {
			result.MemoryAvailable = true
			result.Memory = insertProcessTop(result.Memory, current.item, func(item ProcessTopItem) float64 { return float64(item.MemoryRSS) })
		}
		if current.item.IOAvailable {
			result.IOAvailable = true
			result.IO = insertProcessTop(result.IO, current.item, func(item ProcessTopItem) float64 { return item.ReadRate + item.WriteRate })
		}
	}
	// 名称只为三榜并集最多 15 个 PID 按需读取，按创建时间绑定并缓存一分钟。
	names := make(map[int32]string, 3*processTopLimit)
	for _, items := range [][]ProcessTopItem{result.CPU, result.Memory, result.IO} {
		for index := range items {
			item := &items[index]
			if name, exists := names[item.PID]; exists {
				item.Name = name
				continue
			}
			previous, tracked := s.previous[item.PID]
			name := item.Name
			if (name == "" || now.Sub(previous.nameAt) >= time.Minute) && ctx.Err() == nil {
				p := &process.Process{Pid: item.PID}
				if created, err := p.CreateTimeWithContext(ctx); ctx.Err() == nil && err == nil && created == item.CreateTime {
					value, err := p.NameWithContext(ctx)
					var checkErr error
					if ctx.Err() == nil && err == nil {
						created, checkErr = (&process.Process{Pid: item.PID}).CreateTimeWithContext(ctx)
					}
					if err == nil && checkErr == nil && created == item.CreateTime && value != "" && ctx.Err() == nil {
						name = strings.Map(func(r rune) rune {
							if r < 32 || r == 127 {
								return ' '
							}
							return r
						}, strings.ToValidUTF8(value, "�"))
						if len(name) > 256 {
							name = strings.ToValidUTF8(name[:256], "")
						}
						if tracked && previous.item.CreateTime == item.CreateTime {
							previous.item.Name, previous.nameAt = name, now
							s.previous[item.PID] = previous
						}
					} else {
						result.Partial = true
					}
				} else {
					result.Partial = true
				}
			}
			if name == "" {
				name = "PID " + strconv.Itoa(int(item.PID))
			}
			item.Name, names[item.PID] = name, name
		}
	}
	if ctx.Err() != nil {
		result.Partial, result.Error = true, "进程扫描预算或上下文已结束，下轮继续后续 PID"
	}
	if result.Partial && result.Error == "" {
		result.Error = "部分进程指标正在建立基线，或因权限、退出及扫描预算限制不可读取"
		if runtime.GOOS == "windows" {
			result.Error = "Windows 暂不支持进程磁盘 I/O 排行，CPU 与内存仅统计可读取的进程"
		}
	}
	return result
}

func insertProcessTop(items []ProcessTopItem, item ProcessTopItem, score func(ProcessTopItem) float64) []ProcessTopItem {
	value := score(item)
	if len(items) == processTopLimit {
		last := items[len(items)-1]
		lastValue := score(last)
		if lastValue > value || (lastValue == value && last.PID < item.PID) {
			return items
		}
	}
	index := 0
	for index < len(items) {
		current := score(items[index])
		if current < value || (current == value && items[index].PID >= item.PID) {
			break
		}
		index++
	}
	if index >= processTopLimit {
		return items
	}
	if len(items) < processTopLimit {
		items = append(items, ProcessTopItem{})
	}
	copy(items[index+1:], items[index:len(items)-1])
	items[index] = item
	return items
}
