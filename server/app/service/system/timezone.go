package system

import (
	"server/app/util/platform"
	"time"
)

// GetTimezone 返回本机时区、可选时区和系统当地时间，不使用进程缓存的 time.Local。
func (s *service) GetTimezone() (map[string]interface{}, error) {
	timezones, err := platform.TimezonesList()
	if err != nil {
		return nil, err
	}
	name, now, err := platform.TimezoneGet()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"timezone":  name,
		"timezones": timezones,
		"time":      now.Format(time.RFC3339),
		"timestamp": now.Unix(),
	}, nil
}

// SetTimezone 修改系统时区，不调整系统时间，也不重启已运行的服务。
func (s *service) SetTimezone(name string) error {
	return platform.TimezoneSet(name)
}
