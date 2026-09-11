package daemon

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AutoStartOptions 描述系统自启动项。
// Arguments 为空时默认以前台 run 模式启动，避免再次创建内部守护进程。
type AutoStartOptions struct {
	Name             string
	Executable       string
	WorkingDirectory string
	Arguments        []string
}

// installAutoStart 安装并启用系统自启动。
func (u *Daemon) installAutoStart() error {
	options, err := u.normalizeAutoStartOptions()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		return u.installSystemd(options)
	case "windows":
		return u.installWindowsTask(options)
	case "darwin":
		return u.installLaunchAgent(options)
	default:
		return fmt.Errorf("当前系统不支持安装自启动: %s", runtime.GOOS)
	}
}

// uninstallAutoStart 删除系统自启动。
func (u *Daemon) uninstallAutoStart() error {
	options, err := u.normalizeAutoStartOptions()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		return u.uninstallSystemd(options)
	case "windows":
		return u.uninstallWindowsTask(options)
	case "darwin":
		return u.uninstallLaunchAgent(options)
	default:
		return fmt.Errorf("当前系统不支持删除自启动: %s", runtime.GOOS)
	}
}

func (u *Daemon) normalizeAutoStartOptions() (AutoStartOptions, error) {
	options := u.autoStart
	var err error
	options.Name = strings.TrimSpace(options.Name)
	if options.Name == "" {
		return AutoStartOptions{}, errors.New("自启动名称不能为空")
	}
	for index := range options.Name {
		value := options.Name[index]
		if (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') ||
			(value >= '0' && value <= '9') || value == '-' || value == '_' || value == '.' {
			continue
		}
		return AutoStartOptions{}, errors.New("自启动名称只能包含字母、数字、点、横线和下划线")
	}
	if options.Executable == "" {
		options.Executable, err = os.Executable()
		if err != nil {
			return AutoStartOptions{}, fmt.Errorf("获取程序路径失败: %w", err)
		}
	}
	options.Executable, err = filepath.Abs(options.Executable)
	if err != nil {
		return AutoStartOptions{}, fmt.Errorf("解析程序路径失败: %w", err)
	}
	if options.WorkingDirectory == "" {
		options.WorkingDirectory = filepath.Dir(options.Executable)
	} else {
		options.WorkingDirectory, err = filepath.Abs(options.WorkingDirectory)
		if err != nil {
			return AutoStartOptions{}, fmt.Errorf("解析工作目录失败: %w", err)
		}
	}
	if info, err := os.Stat(options.Executable); err != nil {
		return AutoStartOptions{}, fmt.Errorf("读取程序路径失败: %w", err)
	} else if info.IsDir() {
		return AutoStartOptions{}, errors.New("程序路径不能是目录")
	}
	if info, err := os.Stat(options.WorkingDirectory); err != nil {
		return AutoStartOptions{}, fmt.Errorf("读取工作目录失败: %w", err)
	} else if !info.IsDir() {
		return AutoStartOptions{}, errors.New("工作目录不是目录")
	}
	if len(options.Arguments) == 0 {
		options.Arguments = []string{"run"}
	} else {
		options.Arguments = append([]string(nil), options.Arguments...)
	}
	return options, nil
}

func (u *Daemon) commandOutput(name string, args ...string) error {
	output, err := u.commandText(name, args...)
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return fmt.Errorf("执行 %s 失败: %w: %s", name, err, message)
		}
		return fmt.Errorf("执行 %s 失败: %w", name, err)
	}
	return nil
}

func (u *Daemon) commandText(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func (u *Daemon) quoteSystemd(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return `"` + value + `"`
}

func (u *Daemon) systemdUnitPath(name string) string {
	return filepath.Join("/etc/systemd/system", name+".service")
}

func (u *Daemon) installSystemd(options AutoStartOptions) error {
	arguments := make([]string, 0, len(options.Arguments)+1)
	arguments = append(arguments, u.quoteSystemd(options.Executable))
	for _, argument := range options.Arguments {
		arguments = append(arguments, u.quoteSystemd(argument))
	}
	unit := fmt.Sprintf(`[Unit]
Description=%s
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, u.quoteSystemd(options.Name), u.quoteSystemd(options.WorkingDirectory), strings.Join(arguments, " "))
	path := u.systemdUnitPath(options.Name)
	if err := os.WriteFile(path, []byte(unit), 0644); err != nil {
		return fmt.Errorf("写入 systemd 服务文件失败: %w", err)
	}
	if err := u.commandOutput("systemctl", "daemon-reload"); err != nil {
		return err
	}
	if err := u.commandOutput("systemctl", "enable", options.Name+".service"); err != nil {
		return err
	}
	return nil
}

func (u *Daemon) uninstallSystemd(options AutoStartOptions) error {
	serviceName := options.Name + ".service"
	if _, err := os.Stat(u.systemdUnitPath(options.Name)); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err := u.commandOutput("systemctl", "disable", "--now", serviceName); err != nil {
		return err
	}
	if err := os.Remove(u.systemdUnitPath(options.Name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除 systemd 服务文件失败: %w", err)
	}
	return u.commandOutput("systemctl", "daemon-reload")
}

func (u *Daemon) windowsTaskScriptPath(options AutoStartOptions) string {
	return filepath.Join(options.WorkingDirectory, "."+options.Name+"-startup.cmd")
}

func (u *Daemon) installWindowsTask(options AutoStartOptions) error {
	if err := u.writeWindowsTaskScript(options); err != nil {
		return err
	}
	scriptPath := u.windowsTaskScriptPath(options)
	if err := u.commandOutput("schtasks.exe", "/Create", "/TN", options.Name, "/TR", fmt.Sprintf(`"%s"`, scriptPath), "/SC", "ONSTART", "/RU", "SYSTEM", "/F"); err != nil {
		_ = os.Remove(scriptPath)
		return err
	}
	return nil
}

// writeWindowsTaskScript 保存启动参数，供安装和带参数启动时共用。
func (u *Daemon) writeWindowsTaskScript(options AutoStartOptions) error {
	if strings.ContainsAny(options.WorkingDirectory, "\r\n\"") {
		return errors.New("Windows自启动路径包含不支持的字符")
	}
	arguments := make([]string, 0, len(options.Arguments)+1)
	for _, argument := range append([]string{options.Executable}, options.Arguments...) {
		if strings.ContainsAny(argument, "\r\n\"") {
			return errors.New("Windows自启动参数包含不支持的字符")
		}
		argument = strings.ReplaceAll(argument, "%", "%%")
		argument += strings.Repeat(`\`, len(argument)-len(strings.TrimRight(argument, `\`)))
		arguments = append(arguments, `"`+argument+`"`)
	}
	script := fmt.Sprintf("@echo off\r\nsetlocal DisableDelayedExpansion\r\ncd /d \"%s\"\r\n%s\r\n",
		strings.ReplaceAll(options.WorkingDirectory, "%", "%%"), strings.Join(arguments, " "))
	if err := os.WriteFile(u.windowsTaskScriptPath(options), []byte(script), 0644); err != nil {
		return fmt.Errorf("写入 Windows 自启动脚本失败: %w", err)
	}
	return nil
}

func (u *Daemon) uninstallWindowsTask(options AutoStartOptions) error {
	if err := u.commandOutput("schtasks.exe", "/Query", "/TN", options.Name); err != nil {
		return nil
	}
	if err := u.commandOutput("schtasks.exe", "/Delete", "/TN", options.Name, "/F"); err != nil {
		return err
	}
	if err := os.Remove(u.windowsTaskScriptPath(options)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除 Windows 自启动脚本失败: %w", err)
	}
	return nil
}

func (u *Daemon) launchAgentPath(options AutoStartOptions) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", options.Name+".plist"), nil
}

func (u *Daemon) launchAgentLogPath(options AutoStartOptions) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	return filepath.Join(home, "Library", "Logs", options.Name+".log"), nil
}

func (u *Daemon) xmlText(value string) string {
	var output bytes.Buffer
	_ = xml.EscapeText(&output, []byte(value))
	return output.String()
}

func (u *Daemon) installLaunchAgent(options AutoStartOptions) error {
	plistPath, err := u.launchAgentPath(options)
	if err != nil {
		return err
	}
	logPath, err := u.launchAgentLogPath(options)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("创建 LaunchAgents 目录失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("创建 LaunchAgent 日志目录失败: %w", err)
	}
	var content strings.Builder
	content.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>`)
	content.WriteString("<key>Label</key><string>" + u.xmlText(options.Name) + "</string>")
	content.WriteString("<key>ProgramArguments</key><array>")
	content.WriteString("<string>" + u.xmlText(options.Executable) + "</string>")
	for _, argument := range options.Arguments {
		content.WriteString("<string>" + u.xmlText(argument) + "</string>")
	}
	content.WriteString("</array>")
	content.WriteString("<key>WorkingDirectory</key><string>" + u.xmlText(options.WorkingDirectory) + "</string>")
	content.WriteString("<key>RunAtLoad</key><true/>")
	content.WriteString("<key>KeepAlive</key><true/>")
	content.WriteString("<key>StandardOutPath</key><string>" + u.xmlText(logPath) + "</string>")
	content.WriteString("<key>StandardErrorPath</key><string>" + u.xmlText(logPath) + "</string>")
	content.WriteString("</dict></plist>\n")
	if err := os.WriteFile(plistPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("写入 LaunchAgent 文件失败: %w", err)
	}
	return nil
}

func (u *Daemon) uninstallLaunchAgent(options AutoStartOptions) error {
	path, err := u.launchAgentPath(options)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除 LaunchAgent 文件失败: %w", err)
	}
	return nil
}
