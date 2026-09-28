//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package timezone

import (
	"fmt"
	"runtime"
	"time"
)

// List 返回不支持的平台错误。
func List() ([]string, error) {
	return nil, fmt.Errorf("当前系统暂不支持时区管理: %s", runtime.GOOS)
}

// Get 返回不支持的平台错误。
func Get() (string, time.Time, error) {
	return "", time.Time{}, fmt.Errorf("当前系统暂不支持时区管理: %s", runtime.GOOS)
}

func setTimezone(name string) error {
	return fmt.Errorf("当前系统暂不支持时区管理: %s", runtime.GOOS)
}
