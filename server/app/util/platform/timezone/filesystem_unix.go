//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	localtimePath    = "/etc/localtime"
	timezoneNamePath = "/etc/timezone"
)

// writeTimezone 在系统服务不可用时替换时区链接，并同步发行版的名称记录。
func writeTimezone(name, source string) error {
	if os.Geteuid() != 0 {
		return errors.New("修改系统时区需要 root 权限")
	}
	if err := checkHardwareClock(); err != nil {
		return err
	}
	// 锁覆盖链接与名称记录的整个更新过程，也防止多个面板进程交叉写入。
	directory := filepath.Dir(localtimePath)
	lock, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("锁定系统时区配置失败: %w", err)
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN)
	temporary, err := os.MkdirTemp(directory, ".timezone-")
	if err != nil {
		return fmt.Errorf("准备系统时区失败: %w", err)
	}
	defer os.RemoveAll(temporary)
	link := filepath.Join(temporary, "localtime")
	if err := os.Symlink(source, link); err != nil {
		return err
	}
	// 替换链接本身，避免写入共享 TZif；macOS 的 notifyd 会监听此路径变化。
	if err := os.Rename(link, localtimePath); err != nil {
		return fmt.Errorf("设置系统时区失败: %w", err)
	}
	paths := []string{timezoneNamePath}
	if runtime.GOOS == "freebsd" || runtime.GOOS == "dragonfly" {
		paths = append(paths, "/var/db/zoneinfo")
	}
	for _, path := range paths {
		if err := writeTimezoneName(path, name); err != nil {
			return fmt.Errorf("系统时区已修改，但同步 %s 失败: %w", path, err)
		}
	}
	return nil
}

// checkHardwareClock 本地时间 RTC 还需调整内核偏移，不能仅改文件就报告成功。
func checkHardwareClock() error {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/etc/adjtime")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("读取硬件时钟配置失败: %w", err)
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) >= 3 && strings.TrimSpace(lines[2]) == "LOCAL" {
			return errors.New("硬件时钟使用本地时间，修改时区需要可用的系统时区服务")
		}
	case "freebsd", "dragonfly":
		if _, err := os.Stat("/etc/wall_cmos_clock"); err == nil {
			return errors.New("硬件时钟使用本地时间，当前系统暂不支持直接修改时区")
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// writeTimezoneName 原子更新名称并保留原权限；仅同步已有的 /etc/timezone。
func writeTimezoneName(path, name string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) && path == timezoneNamePath {
		return nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if data, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(data)) == name {
		return nil
	}
	mode := os.FileMode(0644)
	if info != nil && info.Mode().IsRegular() {
		mode = info.Mode().Perm()
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".timezone-")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	_, writeErr := file.WriteString(name + "\n")
	if err := errors.Join(writeErr, file.Chmod(mode), file.Close()); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
