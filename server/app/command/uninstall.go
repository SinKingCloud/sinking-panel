package command

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"server/app/constant"
)

// uninstall 在删除软件前要求明确确认，避免误操作。
func (s *Server) uninstall() error {
	fmt.Print("将停止服务、删除自启动及面板数据，输入 yes 确认卸载: ")
	confirmation, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && strings.TrimSpace(confirmation) == "" {
		fmt.Println("已取消卸载")
		return nil
	}
	if strings.TrimSpace(confirmation) != "yes" {
		fmt.Println("已取消卸载")
		return nil
	}
	if err := s.daemon.Uninstall(); err != nil {
		return fmt.Errorf("停止服务或卸载自启动失败: %w", err)
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return fmt.Errorf("解析程序路径失败: %w", err)
	}
	root := filepath.Dir(executable)
	if root == filepath.Dir(root) {
		return errors.New("拒绝删除文件系统根目录")
	}
	if home, homeErr := os.UserHomeDir(); homeErr == nil {
		home, compareErr := filepath.Abs(home)
		if compareErr != nil {
			return fmt.Errorf("校验用户目录失败: %w", compareErr)
		}
		if filepath.Clean(root) == filepath.Clean(home) {
			return errors.New("拒绝删除用户主目录")
		}
	}

	files := []string{
		executable,
		filepath.Join(root, pidFileName),
		filepath.Join(root, logFileName),
	}
	directoryValues := []string{constant.TempPath, constant.DBPath, constant.ConfPath}
	directories := make([]string, 0, len(directoryValues))
	seenDirectories := make(map[string]struct{}, len(directoryValues))
	for _, directoryValue := range directoryValues {
		directory, err := filepath.Abs(directoryValue)
		if err != nil {
			return fmt.Errorf("解析卸载目录失败: %w", err)
		}
		directory = filepath.Clean(directory)
		relative, containmentErr := filepath.Rel(root, directory)
		if containmentErr != nil {
			return fmt.Errorf("校验卸载目录失败: %w", containmentErr)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("卸载目录不在程序安装目录内")
		}
		if filepath.Clean(directory) == filepath.Clean(root) {
			return errors.New("拒绝删除安装目录本身")
		}
		if _, ok := seenDirectories[directory]; ok {
			continue
		}
		seenDirectories[directory] = struct{}{}
		directories = append(directories, directory)
	}
	if runtime.GOOS == "windows" {
		for _, path := range append(append([]string(nil), files...), directories...) {
			if strings.ContainsAny(path, "\r\n\"") {
				return errors.New("卸载路径包含不支持的字符")
			}
		}
		scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf(".%s-uninstall-%d.cmd", serviceName, os.Getpid()))
		var script strings.Builder
		script.WriteString("@echo off\r\n")
		script.WriteString("timeout /t 1 /nobreak >nul\r\n")
		for _, path := range files {
			script.WriteString(fmt.Sprintf("if exist \"%s\" del /f /q \"%s\"\r\n", path, path))
		}
		for _, path := range directories {
			script.WriteString(fmt.Sprintf("if exist \"%s\" rmdir /s /q \"%s\"\r\n", path, path))
		}
		script.WriteString("del /f /q \"%~f0\"\r\n")
		if err := os.WriteFile(scriptPath, []byte(script.String()), 0600); err != nil {
			return fmt.Errorf("创建 Windows 卸载脚本失败: %w", err)
		}
		if err := exec.Command("cmd.exe", "/C", "start", "", "/B", scriptPath).Start(); err != nil {
			_ = os.Remove(scriptPath)
			return fmt.Errorf("启动 Windows 卸载脚本失败: %w", err)
		}
		fmt.Println("卸载清理将在程序退出后完成")
		return nil
	}
	var removalErrors []error
	for _, path := range files {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			removalErrors = append(removalErrors, fmt.Errorf("删除 %s 失败: %w", path, err))
		}
	}
	for _, path := range directories {
		if err := os.RemoveAll(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			removalErrors = append(removalErrors, fmt.Errorf("删除 %s 失败: %w", path, err))
		}
	}
	if err := errors.Join(removalErrors...); err != nil {
		return fmt.Errorf("删除面板文件失败: %w", err)
	}
	fmt.Println("卸载完成")
	return nil
}
