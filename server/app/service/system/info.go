package system

// GetInfo 获取系统信息
func (s *service) GetInfo() map[string]interface{} {
	// 从各个独立缓存中读取并组装系统信息
	result := make(map[string]interface{})

	// 获取系统基本信息
	s.systemBaseCacheLock.RLock()
	result["system"] = s.systemBaseCache
	s.systemBaseCacheLock.RUnlock()

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

	// 获取运行时信息
	s.runtimeInfoCacheLock.RLock()
	result["runtime"] = s.runtimeInfoCache
	s.runtimeInfoCacheLock.RUnlock()

	// 获取网卡信息
	s.networkInfoCacheLock.RLock()
	result["network"] = s.networkInfoCache
	s.networkInfoCacheLock.RUnlock()

	return result
}
