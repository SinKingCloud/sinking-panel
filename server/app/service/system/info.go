package system

// GetInfo 获取系统信息
func (s *Service) GetInfo() map[string]interface{} {
	// 从缓存中读取系统信息
	systemInfoCacheLock.RLock()
	data := systemInfoCache
	systemInfoCacheLock.RUnlock()
	return data
}
