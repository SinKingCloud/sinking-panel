package system

// GetInfo 获取系统信息
func (s *Service) GetInfo() map[string]interface{} {
	// 从各个独立缓存中读取并组装系统信息
	result := make(map[string]interface{})

	// 获取系统基本信息
	systemBaseCacheLock.RLock()
	result["system"] = systemBaseCache
	systemBaseCacheLock.RUnlock()

	// 获取CPU信息
	cpuInfoCacheLock.RLock()
	result["cpu"] = cpuInfoCache
	cpuInfoCacheLock.RUnlock()

	// 获取内存信息
	memoryInfoCacheLock.RLock()
	result["memory"] = memoryInfoCache
	memoryInfoCacheLock.RUnlock()

	// 获取磁盘信息
	disksInfoCacheLock.RLock()
	result["disks"] = disksInfoCache
	disksInfoCacheLock.RUnlock()

	// 获取系统负载信息
	loadInfoCacheLock.RLock()
	result["load"] = loadInfoCache
	loadInfoCacheLock.RUnlock()

	// 获取运行时信息
	runtimeInfoCacheLock.RLock()
	result["runtime"] = runtimeInfoCache
	runtimeInfoCacheLock.RUnlock()

	// 获取网卡信息
	networkInfoCacheLock.RLock()
	result["network"] = networkInfoCache
	networkInfoCacheLock.RUnlock()

	return result
}
