package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	goDaemon "github.com/sevlyar/go-daemon"

	"server/app/util/ipc"
)

const defaultStopTimeout = 2 * time.Minute

type Daemon struct {
	ControlPath string
	LogFileName string
	StopTimeout time.Duration
	childArgs   []string
	restarting  atomic.Bool
	autoStart   AutoStartOptions
	Service     func(<-chan struct{})
	mu          sync.Mutex
	address     string
}

// SetChildArgs 设置守护进程子进程参数。
func (u *Daemon) SetChildArgs(args ...string) *Daemon {
	u.childArgs = append([]string(nil), args...)
	return u
}

// SetAutoStartOptions 设置系统自启动配置。
func (u *Daemon) SetAutoStartOptions(options AutoStartOptions) *Daemon {
	u.autoStart = options
	if options.Arguments != nil {
		u.autoStart.Arguments = append([]string{}, options.Arguments...)
	}
	return u
}

// InstallAutoStart 安装并启用系统自启动。
func (u *Daemon) InstallAutoStart() error { return u.installAutoStart() }

// UninstallAutoStart 删除系统自启动。
func (u *Daemon) UninstallAutoStart() error { return u.uninstallAutoStart() }

// IsChildProcess 是否为守护进程子进程。
func (u *Daemon) IsChildProcess() bool { return goDaemon.WasReborn() }

// NewDaemon 实例化跨平台进程守护，controlPath 为本机通信目录，不存储 PID。
func NewDaemon(controlPath, logFileName string, service func(<-chan struct{})) (*Daemon, error) {
	if controlPath == "" || logFileName == "" || service == nil {
		return nil, errors.New("参数不能为空")
	}
	path, err := filepath.Abs(controlPath)
	if err != nil {
		return nil, err
	}
	return &Daemon{
		ControlPath: path,
		LogFileName: logFileName,
		StopTimeout: defaultStopTimeout,
		Service:     service,
	}, nil
}

// Running 查询本机控制端点；初始化和退出清理期间仍视为运行中。
func (u *Daemon) Running() (bool, error) {
	_, err := u.control("status", 5*time.Second)
	if errors.Is(err, ipc.ErrUnavailable) {
		return false, nil
	}
	return err == nil, err
}

// Run 在初始化业务前取得独占端点，业务清理完成后才释放。
func (u *Daemon) Run() error {
	listener, err := ipc.Listen(u.ControlPath)
	if err != nil {
		return fmt.Errorf("服务已经运行或控制端点不可用: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	u.mu.Lock()
	u.address = listener.Addr().String()
	u.mu.Unlock()
	done := make(chan struct{})
	closed := make(chan struct{})
	go u.serve(listener, stop, done, closed)
	defer func() {
		// 已接收的 stop 请求在业务资源释放后得到确认。
		_ = listener.Close()
		close(done)
		<-closed
		u.mu.Lock()
		u.address = ""
		u.mu.Unlock()
	}()
	u.Service(ctx.Done())
	return nil
}

// Start 使用已安装的系统服务或独立子进程启动，不创建 PID 文件。
func (u *Daemon) Start() error {
	if u.IsChildProcess() {
		return u.Run()
	}
	if running, err := u.Running(); err != nil {
		return err
	} else if running {
		return errors.New("服务已经运行")
	}
	if u.systemdInstalled() {
		return u.commandOutput("systemctl", "start", u.autoStart.Name+".service")
	}
	options, err := u.normalizeOptions()
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" && options.Name != "" {
		path, err := u.launchAgentPath(options)
		if err != nil {
			return err
		}
		if _, err = os.Stat(path); err == nil {
			uid := strconv.Itoa(os.Getuid())
			var failures []error
			domains := []string{"gui/" + uid, "user/" + uid}
			if os.Getuid() == 0 {
				domains = []string{"system"}
			}
			for _, domain := range domains {
				target := domain + "/" + options.Name
				if _, err := u.commandText("launchctl", "print", target); err != nil {
					if err = u.commandOutput("launchctl", "bootstrap", domain, path); err != nil {
						failures = append(failures, err)
						continue
					}
				}
				return u.commandOutput("launchctl", "kickstart", target)
			}
			return errors.Join(failures...)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if runtime.GOOS == "windows" && options.Name != "" {
		if _, err := os.Stat(u.windowsTaskScriptPath(options)); err == nil {
			if err = u.writeWindowsTaskScript(options); err != nil {
				return err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	command := exec.Command(options.Executable, options.Arguments...)
	if len(u.childArgs) > 0 {
		command = exec.Command(u.childArgs[0], u.childArgs[1:]...)
	}
	command.Dir = options.WorkingDirectory
	command.Env = append(os.Environ(), goDaemon.MARK_NAME+"="+goDaemon.MARK_VALUE)
	logPath := u.LogFileName
	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(options.WorkingDirectory, logPath)
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("打开服务日志失败: %w", err)
	}
	defer file.Close()
	command.Stdout, command.Stderr = file, file
	command, err = u.detachRestart(command)
	if err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}
	exited := make(chan error, 1)
	go func() { exited <- command.Wait() }()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		pid, err := u.control("status", time.Second)
		if err == nil {
			if pid != command.Process.Pid {
				return errors.New("服务已被另一进程启动")
			}
			return nil
		}
		select {
		case err := <-exited:
			if err == nil {
				err = errors.New("服务在建立控制端点前退出")
			}
			return fmt.Errorf("启动服务失败，请查看 %s: %w", logPath, err)
		case <-deadline.C:
			return errors.New("等待服务启动超时，请检查服务日志")
		case <-ticker.C:
		}
	}
}

// Stop 通知服务平滑退出并等待清理完成；未运行时直接返回。
func (u *Daemon) Stop() error {
	pid, err := u.control("status", 5*time.Second)
	if err != nil && !errors.Is(err, ipc.ErrUnavailable) {
		return err
	}
	if u.systemdInstalled() {
		if err := u.syncSystemdKillMode(); err != nil {
			return err
		}
		output, err := u.commandText("systemctl", "show", "--property=MainPID", "--value", u.autoStart.Name+".service")
		if err != nil {
			return err
		}
		current, err := strconv.Atoi(strings.TrimSpace(string(output)))
		if err != nil {
			return err
		}
		if current > 0 && pid > 0 && current != pid {
			return errors.New("同名系统服务正在运行其他进程")
		}
		if current > 0 || pid == 0 {
			return u.commandOutput("systemctl", "stop", u.autoStart.Name+".service")
		}
	}
	if pid == 0 {
		return nil
	}
	if runtime.GOOS == "darwin" && u.autoStart.Name != "" {
		uid := strconv.Itoa(os.Getuid())
		for _, domain := range []string{"gui/" + uid, "user/" + uid, "system"} {
			target := domain + "/" + u.autoStart.Name
			output, err := u.commandText("launchctl", "print", target)
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(output), "\n") {
				if strings.TrimSpace(line) == "pid = "+strconv.Itoa(pid) {
					// 卸载已加载任务但保留 plist，避免 KeepAlive 在停止后重拉。
					if err := u.commandOutput("launchctl", "bootout", target); err != nil {
						return err
					}
					timeout := u.StopTimeout
					if timeout <= 0 {
						timeout = defaultStopTimeout
					}
					deadline := time.Now().Add(timeout)
					for time.Now().Before(deadline) {
						current, err := u.control("status", time.Second)
						if errors.Is(err, ipc.ErrUnavailable) {
							return nil
						}
						if err == nil && current != pid {
							return errors.New("服务进程已变更，请重新执行操作")
						}
						time.Sleep(100 * time.Millisecond)
					}
					return errors.New("等待服务停止超时")
				}
			}
		}
	}
	_, err = u.control("stop "+strconv.Itoa(pid), u.StopTimeout)
	if errors.Is(err, ipc.ErrUnavailable) {
		return nil
	}
	return err
}

// Reload 从外部停止服务后重新启动。
func (u *Daemon) Reload() error {
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
		return err
	}
	_ = file.Close()
	if err := u.Stop(); err != nil {
		return fmt.Errorf("无法停止守护进程: %w", err)
	}
	return u.Start()
}
