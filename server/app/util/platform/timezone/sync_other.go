//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package timezone

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func sources(context.Context) ([]string, error) {
	return nil, fmt.Errorf("当前系统暂不支持时间同步: %s", runtime.GOOS)
}

func setTime(time.Time) error {
	return fmt.Errorf("当前系统暂不支持时间同步: %s", runtime.GOOS)
}
