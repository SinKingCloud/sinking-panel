package command

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/moby/sys/mountinfo"

	"server/app/constant"
)

func readUninstallChoices(input io.Reader, output io.Writer) (confirmed, keepContainers bool, err error) {
	reader := bufio.NewReader(input)
	fmt.Fprint(output, "将停止服务并删除自启动、程序及配置，网站数据会保留，输入 yes 确认卸载: ")
	confirmation, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, false, fmt.Errorf("读取卸载确认失败: %w", err)
	}
	if strings.TrimSpace(confirmation) != "yes" {
		fmt.Fprintln(output, "已取消卸载")
		return false, false, nil
	}
	fmt.Fprint(output, "是否保留轻量容器及其数据？输入 yes 保留，其他输入（包括直接回车）将一并删除: ")
	retention, err := reader.ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return false, false, fmt.Errorf("读取容器保留选项失败: %w", err)
		}
		if retention == "" {
			fmt.Fprintln(output, "已取消卸载")
			return false, false, nil
		}
	}
	return true, strings.TrimSpace(retention) == "yes", nil
}

// uninstall 在删除软件前要求明确确认，避免误操作。
func (s *Server) uninstall() error {
	confirmed, keepContainers, err := readUninstallChoices(os.Stdin, os.Stdout)
	if err != nil || !confirmed {
		return err
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
	// 面板配置和数据库位于 data/server，网站文件所在目录保留。
	directoryValues := []string{constant.RuntimePath, constant.PanelDataPath}
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
	if err := s.daemon.Uninstall(); err != nil {
		return fmt.Errorf("停止服务或卸载自启动失败: %w", err)
	}
	if !keepContainers {
		if err := removeUninstallContainers(root); err != nil {
			return fmt.Errorf("清理轻量容器失败，面板文件尚未删除: %w", err)
		}
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
	if keepContainers {
		fmt.Println("卸载完成，轻量容器数据已保留")
	} else {
		fmt.Println("卸载完成，轻量容器及其数据已删除")
	}
	return nil
}

func removeUninstallContainers(root string) error {
	installationRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("解析安装目录失败: %w", err)
	}
	directory := filepath.Join(installationRoot, constant.ContainerPath)
	resolved, err := filepath.EvalSymlinks(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("解析容器目录失败: %w", err)
	}
	if filepath.Clean(resolved) != filepath.Clean(directory) {
		return errors.New("容器目录不能通过符号链接指向其他目录")
	}
	runtimePath := filepath.Join(directory, "runtime")
	if info, err := os.Lstat(runtimePath); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("容器运行状态目录异常，已保留数据")
		}
		entries, err := os.ReadDir(runtimePath)
		if err != nil {
			return fmt.Errorf("读取容器运行状态失败，已保留数据: %w", err)
		}
		if len(entries) != 0 {
			return errors.New("容器仍有残留运行状态，请完成停止清理后重试，已保留数据")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取容器运行状态目录失败，已保留数据: %w", err)
	}
	mounts, err := mountinfo.GetMounts(mountinfo.PrefixFilter(directory))
	if err != nil {
		return fmt.Errorf("读取容器挂载状态失败，已保留数据: %w", err)
	}
	if len(mounts) != 0 {
		return fmt.Errorf("容器数据目录仍有挂载，已保留数据: %s", mounts[0].Mountpoint)
	}
	if err := os.RemoveAll(directory); err != nil {
		return fmt.Errorf("删除容器数据失败: %w", err)
	}
	return nil
}
