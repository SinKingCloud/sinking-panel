package platform

import "server/app/util/platform/timezone"

// ListTimezones 返回当前系统支持的时区名称，按名称排序。
func ListTimezones() ([]string, error) {
	return timezone.List()
}

// GetTimezone 读取系统配置的时区，不受进程 TZ 环境变量或 time.Local 缓存影响。
func GetTimezone() (string, error) {
	return timezone.Get()
}

// SetTimezone 持久化系统时区，不执行外部命令或调整系统时间戳。
func SetTimezone(name string) error {
	return timezone.Set(name)
}
