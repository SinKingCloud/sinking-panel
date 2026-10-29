package system

// GetStatis 获取同一轮缓存中的 CPU、内存、负载和网卡，以及低频刷新的磁盘容量。
func (s *service) GetStatis() map[string]interface{} {
	s.monitorMu.RLock()
	defer s.monitorMu.RUnlock()
	return map[string]interface{}{
		"cpu":     s.resourceCache.CPU,
		"memory":  s.resourceCache.Memory,
		"load":    s.resourceCache.Load,
		"network": s.resourceCache.Network,
		"disks":   s.disksInfoCache,
	}
}
