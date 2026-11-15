package platform

import (
	"context"
	"server/app/util/platform/timezone"
	"time"
)

// TimezonesList 返回当前系统支持的时区名称，按名称排序。
func TimezonesList() ([]string, error) {
	return timezone.List()
}

// TimezoneGet 读取系统配置的时区和当前当地时间，不受进程 TZ 环境变量或 time.Local 缓存影响。
func TimezoneGet() (string, time.Time, error) {
	return timezone.Get()
}

// TimezoneSet 持久化系统时区，不执行外部命令或调整系统时间戳。
func TimezoneSet(name string) error {
	return timezone.Set(name)
}

// TimeSync 使用系统配置的时间源执行一次时间同步，无配置时使用公共 NTP。
func TimeSync(ctx context.Context) error {
	return timezone.Sync(ctx)
}
