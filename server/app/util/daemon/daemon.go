package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
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
		Arguments:        append([]string(nil), options.Arguments...),
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

// Uninstall 停止服务、卸载自启动并删除面板自身文件。
func (u *Daemon) Uninstall() error {
	return u.uninstall()
}

// IsChildProcess 是否为守护进程子进程。
func (u *Daemon) IsChildProcess() bool {
	if runtime.GOOS == "windows" {
		return false
	}
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	u.Service(ctx.Done())
	return nil
}

// Start 启动
func (u *Daemon) Start() error {
	if runtime.GOOS == "windows" {
		return u.startWindows()
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
	// 读取PID文件获取守护进程的PID
	pid, err := u.readPidFile()
	if err != nil {
		return fmt.Errorf("无法读取PID文件: %w", err)
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
	// 先停止守护进程
	if err := u.Stop(); err != nil {
		if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrProcessDone) {
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
