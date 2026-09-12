package system

// GetStatus 获取系统状态信息（网卡流量和磁盘IO）
func (s *service) GetStatus(netInterface, diskName string, after int64) map[string]interface{} {
	s.statusCacheLock.RLock()
	defer s.statusCacheLock.RUnlock()
	if after > s.monitorSequence {
		after = 0
	}

	result := map[string]interface{}{
		"sample_id":   s.monitorSequence,
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
			result["network"] = make(map[string]interface{})
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
			result["disk"] = make(map[string]interface{})
		}
	} else {
		result["disk"] = diskData
	}

	history := make([]map[string]interface{}, 0)
	for _, sample := range s.monitorHistory {
		if sample["sample_id"].(int64) <= after {
			continue
		}

		sampleNetwork := sample["network"].(map[string]map[string]interface{})
		if netInterface != "" {
			if netData, exists := sampleNetwork[netInterface]; exists {
				sampleNetwork = map[string]map[string]interface{}{
					netInterface: netData,
				}
			} else {
				sampleNetwork = make(map[string]map[string]interface{})
			}
		}

		sampleDisk := sample["disk"].(map[string]map[string]interface{})
		if diskName != "" {
			if diskInfo, exists := sampleDisk[diskName]; exists {
				sampleDisk = map[string]map[string]interface{}{
					diskName: diskInfo,
				}
			} else {
				sampleDisk = make(map[string]map[string]interface{})
			}
		}

		history = append(history, map[string]interface{}{
			"sample_id":   sample["sample_id"],
			"sample_time": sample["sample_time"],
			"network":     sampleNetwork,
			"disk":        sampleDisk,
		})
	}
	result["history"] = history

	// 返回完整状态信息
	return result
}
