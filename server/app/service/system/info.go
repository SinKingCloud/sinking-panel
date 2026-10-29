package system

import (
	"os"
	"runtime"
	"server/app/constant"
	"time"
)

var panelStartedAt = time.Now().UnixMilli()

// GetInfo 组合主机信息和面板自身的运行时信息。
func (s *service) GetInfo() map[string]interface{} {
	s.monitorMu.RLock()
	info := s.systemBaseCache
	s.monitorMu.RUnlock()
	return map[string]interface{}{
		"hostname":         info.Hostname,
		"os":               info.OS,
		"platform":         info.Platform,
		"platform_version": info.PlatformVersion,
		"kernel_version":   info.KernelVersion,
		"kernel_arch":      info.KernelArch,
		"uptime":           info.Uptime,
		"boot_time":        info.BootTime,
		"version":          constant.Version,
		"go_version":       runtime.Version(),
		"num_goroutine":    runtime.NumGoroutine(),
		"panel_pid":        os.Getpid(),
		"panel_start_time": panelStartedAt,
	}
}
