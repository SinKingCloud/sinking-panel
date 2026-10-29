package system

import "server/app/util/psutil"

// GetIO 获取网卡和磁盘的 I/O 统计及采样历史。
func (s *service) GetIO(netInterface, diskName string, after int64) map[string]interface{} {
	s.monitorMu.RLock()
	defer s.monitorMu.RUnlock()
	if after > s.monitorSequence {
		after = 0
	}
	filter := func(sample monitorSample) monitorSample {
		if netInterface != "" {
			network := make(map[string]psutil.NetworkIO, 1)
			if value, ok := sample.Network[netInterface]; ok {
				network[netInterface] = value
			}
			sample.Network = network
		}
		if diskName != "" {
			disk := make(map[string]psutil.DiskIO, 1)
			if value, ok := sample.Disk[diskName]; ok {
				disk[diskName] = value
			}
			sample.Disk = disk
		}
		return sample
	}
	current := filter(monitorSample{
		ID:      s.monitorSequence,
		Time:    s.monitorUpdatedAt,
		Network: s.netRateCache,
		Disk:    s.diskRateCache,
	})
	history := make([]monitorSample, 0)
	for _, sample := range s.monitorHistory {
		if sample.ID > after {
			history = append(history, filter(sample))
		}
	}
	return map[string]interface{}{
		"sample_id":   current.ID,
		"sample_time": current.Time,
		"network":     current.Network,
		"disk":        current.Disk,
		"history":     history,
	}
}
