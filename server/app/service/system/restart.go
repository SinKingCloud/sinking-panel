package system

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Restart 先返回请求结果，再执行重启；同一时间只接受一个重启请求。
func (s *service) Restart(action string) error {
	var run func() error
	switch action {
	case "panel":
		if s.daemon == nil {
			return errors.New("面板进程管理器未初始化")
		}
		if executable, err := os.Executable(); err != nil {
			return fmt.Errorf("获取面板程序失败: %w", err)
		} else if _, err = os.Stat(executable); err != nil {
			return fmt.Errorf("读取面板程序失败: %w", err)
		}
		run = s.daemon.Restart
	case "server":
		command, err := serverRestartCommand(runtime.GOOS)
		if err != nil {
			return err
		}
		run = func() error {
			output, err := command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("重启服务器失败: %w: %s", err, strings.TrimSpace(string(output)))
			}
			return nil
		}
	default:
		return errors.New("重启操作无效")
	}
	if s.ctx.Err() != nil {
		return errors.New("面板正在停止")
	}
	if !s.restartPending.CompareAndSwap(false, true) {
		return errors.New("正在重启，请勿重复提交")
	}
	go func() {
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-s.ctx.Done():
			s.restartPending.Store(false)
			return
		case <-timer.C:
		}
		if err := run(); err != nil {
			log.Printf("执行系统重启失败 [%s]: %v", action, err)
			s.restartPending.Store(false)
		}
	}()
	return nil
}

// serverRestartCommand 使用各系统的正常关机流程，仅在命令不存在时选择备用命令。
func serverRestartCommand(goos string) (*exec.Cmd, error) {
	var commands [][]string
	switch goos {
	case "linux":
		if _, err := os.Stat("/run/systemd/system"); err == nil {
			commands = append(commands, []string{"systemctl", "reboot"})
		}
		commands = append(commands, []string{"shutdown", "-r", "now"}, []string{"reboot"})
	case "darwin", "freebsd", "openbsd", "netbsd", "dragonfly":
		commands = [][]string{{"shutdown", "-r", "now"}}
	case "windows":
		commands = [][]string{{"shutdown.exe", "/r", "/t", "0"}}
	case "solaris", "illumos":
		commands = [][]string{{"shutdown", "-y", "-g0", "-i6"}}
	case "aix":
		commands = [][]string{{"shutdown", "-r", "+0"}}
	default:
		return nil, fmt.Errorf("当前系统不支持重启服务器: %s", goos)
	}
	for _, command := range commands {
		paths := []string{command[0]}
		if goos == "windows" {
			if root := os.Getenv("SystemRoot"); filepath.IsAbs(root) {
				paths = append([]string{filepath.Join(root, "System32", command[0])}, paths...)
			}
		} else {
			// 自启动服务的 PATH 可能不包含 sbin。
			for _, directory := range []string{"/sbin", "/usr/sbin", "/usr/bin", "/bin"} {
				paths = append(paths, filepath.Join(directory, command[0]))
			}
		}
		for _, path := range paths {
			if executable, err := exec.LookPath(path); err == nil {
				return exec.Command(executable, command[1:]...), nil
			}
		}
	}
	return nil, fmt.Errorf("找不到服务器重启命令，请检查系统命令是否可用: %s", goos)
}
