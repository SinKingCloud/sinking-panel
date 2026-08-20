package daemon

import (
	"fmt"
	"runtime"
	"strings"
)

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
	if err := u.commandOutput("schtasks.exe", "/Run", "/TN", options.Name); err != nil {
		return fmt.Errorf("启动 Windows 服务任务失败: %w", err)
	}
	return nil
}

// stopWindows 通过计划任务结束前台服务进程。
func (u *Daemon) stopWindows() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前系统不是 Windows")
	}
	options, err := u.normalizeAutoStartOptions()
	if err != nil {
		return err
	}
	if err := u.commandOutput("schtasks.exe", "/Query", "/TN", options.Name); err != nil {
		return nil
	}
	output, err := u.commandText("schtasks.exe", "/Query", "/TN", options.Name, "/FO", "LIST", "/V")
	if err != nil {
		return nil
	}
	status := strings.ToLower(string(output))
	if !strings.Contains(status, "running") && !strings.Contains(status, "运行") {
		return nil
	}
	if err := u.commandOutput("schtasks.exe", "/End", "/TN", options.Name); err != nil {
		return fmt.Errorf("停止 Windows 服务任务失败: %w", err)
	}
	return nil
}
