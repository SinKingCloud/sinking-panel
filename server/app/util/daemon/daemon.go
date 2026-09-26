package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
)

const defaultStopTimeout = 2 * time.Minute

type Daemon struct {
	PidFileName string
	LogFileName string
	StopTimeout time.Duration
	childArgs   []string
	restarting  atomic.Bool
	autoStart   AutoStartOptions
	Service     func(<-chan struct{})
}

// SetChildArgs 设置守护进程子进程参数。
func (u *Daemon) SetChildArgs(args ...string) *Daemon {
	u.childArgs = append([]string(nil), args...)
	return u
}

// SetAutoStartOptions 设置系统自启动配置。
func (u *Daemon) SetAutoStartOptions(options AutoStartOptions) *Daemon {
	u.autoStart = AutoStartOptions{
		Name:             options.Name,
		Executable:       options.Executable,
		WorkingDirectory: options.WorkingDirectory,
		Arguments:        options.Arguments,
	}
	if options.Arguments != nil {
		u.autoStart.Arguments = append([]string{}, options.Arguments...)
	}
	return u
}

// InstallAutoStart 安装并启用系统自启动。
func (u *Daemon) InstallAutoStart() error {
	return u.installAutoStart()
}

// UninstallAutoStart 删除系统自启动。
func (u *Daemon) UninstallAutoStart() error {
	return u.uninstallAutoStart()
}

// Uninstall 停止服务并卸载系统自启动。
func (u *Daemon) Uninstall() error {
	return u.uninstall()
}

// IsChildProcess 是否为守护进程子进程。
func (u *Daemon) IsChildProcess() bool {
	return daemon.WasReborn()
}

// NewDaemon 实例化跨平台进程守护。
func NewDaemon(pidFileName string, logFileName string, service func(<-chan struct{})) (*Daemon, error) {
	if pidFileName == "" || logFileName == "" || service == nil {
		return nil, errors.New("参数不能为空")
	}
	return &Daemon{
		PidFileName: pidFileName,
		LogFileName: logFileName,
		StopTimeout: defaultStopTimeout,
		Service:     service,
	}, nil
}

// Run 在当前进程运行服务，并接收平台停止通知。
func (u *Daemon) Run() error {
	if runtime.GOOS == "windows" {
		return u.runWindows()
	}
	// 系统服务直接运行时也需要 PID 锁，避免与手动启动的进程重复运行。
	if !u.IsChildProcess() {
		pidFile, err := daemon.OpenLockFile(u.PidFileName, 0640)
		if err != nil {
			return fmt.Errorf("打开服务 PID 文件失败: %w", err)
		}
		if err = pidFile.Lock(); err != nil {
			_ = pidFile.Close()
			return fmt.Errorf("服务已经运行: %w", err)
		}
		defer pidFile.Remove()
		if err = pidFile.WritePid(); err != nil {
			return fmt.Errorf("写入服务 PID 失败: %w", err)
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	u.Service(ctx.Done())
	return nil
}

// Start 启动
func (u *Daemon) Start() error {
	if runtime.GOOS == "windows" {
		if u.IsChildProcess() {
			return u.Run()
		}
		return u.startWindows()
	}
	if !u.IsChildProcess() && u.systemdInstalled() {
		return u.commandOutput("systemctl", "start", u.autoStart.Name+".service")
	}
	// 创建守护进程上下文
	daemonCtx := &daemon.Context{
		PidFileName: u.PidFileName,
		LogFileName: u.LogFileName,
		WorkDir:     "./",
		Args:        u.childArgs,
		Umask:       027,
	}
	// 启动守护进程并获取新的进程上下文和PID
	d, err := daemonCtx.Reborn()
	if err != nil {
		return fmt.Errorf("创建守护进程失败: %w", err)
	}
	// 父进程，已经启动了守护进程，直接返回
	if d != nil {
		return nil
	}
	// 子进程同步运行服务，确保服务完成清理后守护进程才退出。
	defer func() {
		_ = daemonCtx.Release()
	}()
	return u.Run()
}

// Stop 停止
func (u *Daemon) Stop() error {
	if runtime.GOOS == "windows" {
		return u.stopWindows()
	}
	managed := u.systemdInstalled()
	if managed {
		if err := u.commandOutput("systemctl", "stop", u.autoStart.Name+".service"); err != nil {
			return err
		}
	}
	// 读取PID文件获取守护进程的PID
	pid, err := u.readPidFile()
	if err != nil {
		if managed && errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("无法读取PID文件: %w", err)
	}
	if pid <= 1 {
		return errors.New("服务 PID 无效")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("无法获取守护进程: %w", err)
	}
	defer func() {
		_ = process.Release()
	}()
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("无法停止守护进程: %w", err)
	}
	deadline := time.Now().Add(u.StopTimeout)
	for {
		// 服务会在资源清理完成后删除 PID 文件；不必等待 systemd 回收进程。
		if _, statErr := os.Stat(u.PidFileName); errors.Is(statErr, os.ErrNotExist) {
			return nil
		}
		err = process.Signal(syscall.Signal(0))
		if err != nil {
			if errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH) {
				break
			}
			return fmt.Errorf("确认守护进程状态失败: %w", err)
		}
		if time.Now().After(deadline) {
			return errors.New("等待守护进程停止超时")
		}
		time.Sleep(100 * time.Millisecond)
	}
	// 删除PID文件
	if err = os.Remove(u.PidFileName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("无法删除PID文件: %w", err)
	}
	return nil
}

// Reload 重启
func (u *Daemon) Reload() error {
	// 启动路径、日志和任务脚本先检查，避免停止服务后才发现无法重新启动。
	options, err := u.normalizeOptions()
	if err != nil {
		return err
	}
	logPath := u.LogFileName
	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(options.WorkingDirectory, logPath)
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("打开服务日志失败: %w", err)
	}
	_ = file.Close()
	if runtime.GOOS == "windows" && options.Name != "" {
		if _, err := os.Stat(u.windowsTaskScriptPath(options)); err == nil {
			if err = u.writeWindowsTaskScript(options); err != nil {
				return err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("检查 Windows 自启动脚本失败: %w", err)
		}
	}
	// 先停止守护进程
	if err := u.Stop(); err != nil {
		if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
			return fmt.Errorf("无法停止守护进程: %w", err)
		}
	}
	time.Sleep(time.Second)
	// 再启动守护进程
	if err := u.Start(); err != nil {
		return fmt.Errorf("无法重新启动守护进程: %w", err)
	}
	return nil
}
