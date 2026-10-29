package system

import "server/app/util/psutil"

// updateMonitorDevices 低频刷新汇总设备，避免每秒遍历系统设备目录。
func (s *service) updateMonitorDevices(diskReady bool) {
	s.monitorMu.RLock()
	diskNames := make([]string, 0, len(s.diskRateCache))
	for name := range s.diskRateCache {
		diskNames = append(diskNames, name)
	}
	networkNames := make([]string, 0, len(s.netRateCache))
	for name := range s.netRateCache {
		networkNames = append(networkNames, name)
	}
	s.monitorMu.RUnlock()
	diskNames, networkNames = psutil.AggregateDevices(diskNames, networkNames)
	disks := make(map[string]bool, len(diskNames))
	for _, name := range diskNames {
		disks[name] = true
	}
	networks := make(map[string]bool, len(networkNames))
	for _, name := range networkNames {
		networks[name] = true
	}
	s.monitorMu.Lock()
	s.monitorDisks = disks
	s.monitorNetworks = networks
	s.monitorDiskReady = diskReady
	s.monitorMu.Unlock()
}

// monitorDiskCapacity 同一文件系统的多次挂载只计算一次。
func (s *service) monitorDiskCapacity(disks []psutil.DiskInfo) (used, total float64) {
	seen := make(map[string]bool, len(disks))
	for _, disk := range disks {
		key := disk.Filesystem
		if key == "" {
			key = disk.Path
		}
		if seen[key] || disk.Size.Total <= 0 {
			continue
		}
		seen[key] = true
		total += float64(disk.Size.Total)
		used += float64(max(0, min(disk.Size.Used, disk.Size.Total)))
	}
	return used, total
}
