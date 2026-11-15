//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// setTime 修改系统实时时钟；由内核检查 root 或 CAP_SYS_TIME 权限。
func setTime(value time.Time) error {
	seconds, err := unix.TimeToTimespec(value)
	if err != nil {
		return fmt.Errorf("时间超出当前系统支持范围: %w", err)
	}
	// 分开转换秒与微秒，避免 UnixNano 溢出或微秒取整改变秒数。
	timestamp := unix.NsecToTimeval(int64(value.Nanosecond()/1000) * 1000)
	timestamp.Sec = seconds.Sec
	return unix.Settimeofday(&timestamp)
}
