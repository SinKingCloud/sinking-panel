package system

// GetStatis 获取 CPU、内存、负载、磁盘和网卡的资源统计。
func (s *service) GetStatis() map[string]interface{} {
	// 从各个独立缓存中读取并组装资源信息
	result := make(map[string]interface{})

	// 获取CPU信息
	s.cpuInfoCacheLock.RLock()
	result["cpu"] = s.cpuInfoCache
	s.cpuInfoCacheLock.RUnlock()

	// 获取内存信息
	s.memoryInfoCacheLock.RLock()
	result["memory"] = s.memoryInfoCache
	s.memoryInfoCacheLock.RUnlock()

	// 获取磁盘信息
	s.disksInfoCacheLock.RLock()
	result["disks"] = s.disksInfoCache
	s.disksInfoCacheLock.RUnlock()

	// 获取系统负载信息
	s.loadInfoCacheLock.RLock()
	result["load"] = s.loadInfoCache
	s.loadInfoCacheLock.RUnlock()

	// 获取网卡信息
	s.networkInfoCacheLock.RLock()
	result["network"] = s.networkInfoCache
	s.networkInfoCacheLock.RUnlock()

	return result
}
