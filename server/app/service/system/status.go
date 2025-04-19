package system

// GetStatus 获取系统状态信息（网卡流量和磁盘IO）
func (s *Service) GetStatus(netInterface, diskName string) map[string]interface{} {
	// 从缓存中获取状态信息
	systemStatusCacheLock.RLock()
	statusData := systemStatusCache
	systemStatusCacheLock.RUnlock()

	// 如果指定了网卡，只返回该网卡的信息
	if netInterface != "" {
		if networkData, ok := statusData["network"].(map[string]map[string]interface{}); ok {
			if netData, exists := networkData[netInterface]; exists {
				return map[string]interface{}{
					"network": map[string]interface{}{
						netInterface: netData,
					},
					"disk": statusData["disk"],
				}
			}
		}
	}

	// 如果指定了磁盘，只返回该磁盘的信息
	if diskName != "" {
		if diskData, ok := statusData["disk"].(map[string]map[string]interface{}); ok {
			if diskInfo, exists := diskData[diskName]; exists {
				return map[string]interface{}{
					"network": statusData["network"],
					"disk": map[string]interface{}{
						diskName: diskInfo,
					},
				}
			}
		}
	}

	// 返回完整状态信息
	return statusData
}
