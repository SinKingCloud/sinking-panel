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
	"server/app/util/platform"
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
	fmt.Fprint(output, "是否保留容器、镜像及数据？输入 yes 保留，重新安装后可继续使用；其他输入（包括直接回车）将全部清理: ")
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
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("解析程序路径失败: %w", err)
	}
	root := filepath.Dir(executable)

	files := []string{
		executable,
		filepath.Join(root, pidFileName),
		filepath.Join(root, pidFileName+".stop"),
		filepath.Join(root, logFileName),
	}
	// 面板配置和数据库位于 data/server，网站文件所在目录保留。
	directories := []string{
		filepath.Join(root, constant.RuntimePath),
		filepath.Join(root, constant.PanelDataPath),
	}
	if err := s.stop(true); err != nil {
		return err
	}
	if err := s.daemon.UninstallAutoStart(); err != nil {
		return fmt.Errorf("卸载系统服务失败: %w", err)
	}
	if !keepContainers {
		directory := filepath.Join(root, constant.ContainerPath)
		// StopAll 已清理容器挂载；确认没有残留挂载再删除，避免触及外部数据。
		if platform.IsLinux() {
			mounts, err := mountinfo.GetMounts(mountinfo.PrefixFilter(directory))
			if err != nil {
				return fmt.Errorf("读取容器挂载失败: %w", err)
			}
			if len(mounts) > 0 {
				return fmt.Errorf("容器目录仍有挂载，已保留数据: %s", mounts[0].Mountpoint)
			}
		}
		if err := os.RemoveAll(directory); err != nil {
			return fmt.Errorf("清理容器、镜像及数据失败，面板文件尚未删除: %w", err)
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
		fmt.Println("卸载完成，容器、镜像及数据已保留，重新安装到原目录后可继续使用")
	} else {
		fmt.Println("卸载完成，容器、镜像及数据已清理")
	}
	return nil
}
