package system

// GetStatus 获取系统状态信息（网卡流量和磁盘IO）
func (s *service) GetStatus(netInterface, diskName string) map[string]interface{} {
	s.statusCacheLock.RLock()
	defer s.statusCacheLock.RUnlock()

	result := map[string]interface{}{
		"sample_time": s.monitorUpdatedAt,
	}

	networkData := s.netRateCache
	diskData := s.diskRateCache

	// 如果指定了网卡，只返回该网卡的信息
	if netInterface != "" {
		if netData, exists := networkData[netInterface]; exists {
			result["network"] = map[string]interface{}{
				netInterface: netData,
			}
		} else {
			result["network"] = map[string]interface{}{}
		}
	} else {
		result["network"] = networkData
	}

	// 如果指定了磁盘，只返回该磁盘的信息
	if diskName != "" {
		if diskInfo, exists := diskData[diskName]; exists {
			result["disk"] = map[string]interface{}{
				diskName: diskInfo,
			}
		} else {
			result["disk"] = map[string]interface{}{}
		}
	} else {
		result["disk"] = diskData
	}

	// 返回完整状态信息
	return result
}
