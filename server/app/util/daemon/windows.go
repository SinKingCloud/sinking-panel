package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func (u *Daemon) runWindows() (result error) {
	signalContext, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignal()
	stopFile := u.PidFileName + ".stop"
	_ = os.Remove(stopFile)
	if err := os.WriteFile(u.PidFileName, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		return fmt.Errorf("写入服务 PID 失败: %w", err)
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	group := sync.WaitGroup{}
	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-signalContext.Done():
				close(stop)
				return
			case <-ticker.C:
				if _, err := os.Stat(stopFile); err == nil {
					close(stop)
					return
				}
			}
		}
	}()
	defer func() {
		close(done)
		group.Wait()
		_ = os.Remove(stopFile)
		if err := os.Remove(u.PidFileName); result == nil && err != nil && !errors.Is(err, os.ErrNotExist) {
			result = fmt.Errorf("删除服务 PID 失败: %w", err)
		}
	}()
	u.Service(stop)
	return nil
}

// startWindows 通过计划任务启动前台服务进程。
func (u *Daemon) startWindows() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前系统不是 Windows")
	}
	options, err := u.normalizeAutoStartOptions()
	if err != nil {
		return err
	}
	if err := u.commandOutput("schtasks.exe", "/Query", "/TN", options.Name); err != nil {
		return fmt.Errorf("Windows 自启动任务不存在，请先执行 install: %w", err)
	}
	if len(options.Arguments) > 1 {
		if err := u.writeWindowsTaskScript(options); err != nil {
			return err
		}
	}
	if err := u.commandOutput("schtasks.exe", "/Run", "/TN", options.Name); err != nil {
		return fmt.Errorf("启动 Windows 服务任务失败: %w", err)
	}
	return nil
}

// stopWindows 通知服务进程自行平滑退出。
func (u *Daemon) stopWindows() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前系统不是 Windows")
	}
	if _, err := os.Stat(u.PidFileName); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("读取服务 PID 失败: %w", err)
	}
	stopFile := u.PidFileName + ".stop"
	if err := os.WriteFile(stopFile, nil, 0600); err != nil {
		return fmt.Errorf("发送停止通知失败: %w", err)
	}
	deadline := time.NewTimer(u.StopTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-deadline.C:
			return errors.New("等待服务停止超时")
		case <-ticker.C:
			if _, err := os.Stat(u.PidFileName); errors.Is(err, os.ErrNotExist) {
				return nil
			} else if err != nil {
				return fmt.Errorf("确认服务状态失败: %w", err)
			}
		}
	}
}
