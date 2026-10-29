package psutil

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// Processes 返回按 PID 升序的进程摘要。枚举期间退出的进程会被跳过，字段权限错误留在 Errors。
func Processes(ctx context.Context) ([]ProcessSummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]ProcessSummary, 0, len(pids))
	for _, pid := range pids {
		if err := ctx.Err(); err != nil {
			return items, err
		}
		p, err := process.NewProcessWithContext(ctx, pid)
		if errors.Is(err, process.ErrorProcessNotRunning) {
			continue
		}
		if err != nil {
			items = append(items, ProcessSummary{
				PID:    pid,
				Errors: map[string]string{"process": err.Error()},
			})
			continue
		}
		items = append(items, processSummary(ctx, p))
	}
	return items, ctx.Err()
}

// NewProcess 绑定当前 PID 对应的进程，后续操作前核对创建时间，避免操作已复用的 PID。
func NewProcess(ctx context.Context, pid int32) (*Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if pid <= 0 {
		return nil, errors.New("进程 ID 必须大于 0")
	}
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, err
	}
	created, err := p.CreateTimeWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if created <= 0 {
		return nil, errors.New("无法确认进程创建时间")
	}
	return &Process{
		pid:        pid,
		createTime: created,
	}, ctx.Err()
}

// Info 获取进程详情。可选字段失败不丢弃整个详情，原因写入 Errors。
func (p *Process) Info(ctx context.Context) (ProcessInfo, error) {
	data := ProcessInfo{
		Command: []string{},
		Status:  []string{},
	}
	current, err := p.open(ctx)
	if err != nil {
		return data, err
	}
	data.ProcessSummary = processSummary(ctx, current)
	record := func(field string, err error) {
		if err != nil {
			data.Errors[field] = err.Error()
		}
	}
	data.Executable, err = current.ExeWithContext(ctx)
	record("executable", err)
	command, err := current.CmdlineSliceWithContext(ctx)
	record("command", err)
	if command != nil {
		data.Command = command
	}
	data.Directory, err = current.CwdWithContext(ctx)
	record("directory", err)
	data.Username, err = current.UsernameWithContext(ctx)
	record("username", err)
	status, err := current.StatusWithContext(ctx)
	record("status", err)
	if status != nil {
		data.Status = status
	}
	_, err = p.open(ctx)
	return data, err
}

// Sample 同步获取资源，速率与此句柄上次成功采样比较。首次只有累计值，速率可用标志为 false。
func (p *Process) Sample(ctx context.Context) (ProcessUsage, error) {
	if p == nil {
		return ProcessUsage{}, errors.New("请通过 NewProcess 创建进程句柄")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	data := ProcessUsage{
		PID:        p.pid,
		CreateTime: p.createTime,
		Errors:     map[string]string{},
	}
	current, err := p.open(ctx)
	if err != nil {
		return data, err
	}
	now := time.Now()
	data.SampleTime = now.UnixMilli()
	if times, err := current.TimesWithContext(ctx); err != nil {
		data.Errors["cpu"] = err.Error()
	} else {
		data.CPUAvailable = true
		data.CPUUser, data.CPUSystem = times.User, times.System
		total := times.User + times.System
		if elapsed := now.Sub(p.cpuAt).Seconds(); !p.cpuAt.IsZero() && elapsed > 0 && total >= p.cpuTime {
			data.CPURateAvailable = true
			data.CPUPercent = (total - p.cpuTime) / elapsed * 100
		}
		p.cpuTime, p.cpuAt = total, now
	}
	if memory, err := current.MemoryInfoWithContext(ctx); err != nil {
		data.Errors["memory"] = err.Error()
	} else {
		data.MemoryAvailable = true
		data.MemoryRSS, data.MemoryVMS, data.MemorySwap = memory.RSS, memory.VMS, memory.Swap
	}
	data.Threads, err = current.NumThreadsWithContext(ctx)
	if err != nil {
		data.Errors["threads"] = err.Error()
	}
	if io, err := current.IOCountersWithContext(ctx); err != nil {
		data.Errors["io"] = err.Error()
	} else {
		data.IOAvailable = true
		data.ReadBytes, data.WriteBytes = io.ReadBytes, io.WriteBytes
		if runtime.GOOS == "darwin" {
			data.ReadBytes, data.WriteBytes = io.DiskReadBytes, io.DiskWriteBytes
		}
		data.ReadCount, data.WriteCount = io.ReadCount, io.WriteCount
		if elapsed := now.Sub(p.ioAt).Seconds(); !p.ioAt.IsZero() && elapsed > 0 && data.ReadBytes >= p.readBytes && data.WriteBytes >= p.writeBytes {
			data.IORateAvailable = true
			data.ReadRate = float64(data.ReadBytes-p.readBytes) / elapsed
			data.WriteRate = float64(data.WriteBytes-p.writeBytes) / elapsed
		}
		p.readBytes, p.writeBytes, p.ioAt = data.ReadBytes, data.WriteBytes, now
	}
	_, err = p.open(ctx)
	return data, err
}

// Control 控制当前进程：terminate（请求结束）、kill（强制结束）、suspend（暂停）、resume（继续）。
// Windows 的 terminate 与 kill 都通过 TerminateProcess 强制结束；不递归处理子进程，不等待退出。
func (p *Process) Control(ctx context.Context, action string) error {
	switch action {
	case "terminate", "kill", "suspend", "resume":
	default:
		return errors.New("进程操作必须为 terminate、kill、suspend 或 resume")
	}
	current, err := p.open(ctx)
	if err != nil {
		return err
	}
	switch action {
	case "terminate":
		err = current.TerminateWithContext(ctx)
	case "kill":
		err = current.KillWithContext(ctx)
	case "suspend":
		err = current.SuspendWithContext(ctx)
	case "resume":
		err = current.ResumeWithContext(ctx)
	}
	return err
}

// open 每次重新获取进程身份，不复用底层对象缓存的创建时间。
func (p *Process) open(ctx context.Context) (*process.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p == nil || p.pid <= 0 || p.createTime <= 0 {
		return nil, errors.New("请通过 NewProcess 创建进程句柄")
	}
	current, err := process.NewProcessWithContext(ctx, p.pid)
	if err != nil {
		return nil, err
	}
	created, err := current.CreateTimeWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if created != p.createTime {
		return nil, fmt.Errorf("进程 %d 已退出，PID 已被其他进程使用", p.pid)
	}
	return current, ctx.Err()
}

func processSummary(ctx context.Context, p *process.Process) ProcessSummary {
	data := ProcessSummary{
		PID:    p.Pid,
		Errors: map[string]string{},
	}
	record := func(field string, err error) {
		if err != nil {
			data.Errors[field] = err.Error()
		}
	}
	var err error
	data.Name, err = p.NameWithContext(ctx)
	record("name", err)
	data.ParentPID, err = p.PpidWithContext(ctx)
	record("parent_pid", err)
	data.CreateTime, err = p.CreateTimeWithContext(ctx)
	record("create_time", err)
	return data
}
