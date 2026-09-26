package system

import (
	"os"
	"runtime"
	"server/app/constant"
	"time"
)

var panelStartedAt = time.Now().UnixMilli()

// GetInfo 获取主机、操作系统和运行时的基础信息。
func (s *service) GetInfo() map[string]interface{} {
	result := map[string]interface{}{
		"version":          constant.Version,
		"go_version":       runtime.Version(),
		"num_goroutine":    runtime.NumGoroutine(),
		"panel_pid":        os.Getpid(),
		"panel_start_time": panelStartedAt,
	}
	// 复制基础信息，避免将运行时字段写入共享缓存。
	s.systemBaseCacheLock.RLock()
	for key, value := range s.systemBaseCache {
		result[key] = value
	}
	s.systemBaseCacheLock.RUnlock()
	return result
}
