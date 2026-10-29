package system

import (
	"errors"
	"time"

	"server/app/util/psutil"
)

// startMonitor 启动系统监控
func (s *service) startMonitor() {
	// 首次同步更新，确保服务启动后即可读取监控数据
	s.updateMonitor()
	s.updateStaticMonitor()
	s.startMonitorHistory()

	s.wait.Add(2)
	go func() {
		defer s.wait.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateMonitor()
			case <-s.ctx.Done():
				return
			}
		}
	}()

	// 磁盘容量等静态信息单独刷新，避免慢挂载点阻塞实时采样
	go func() {
		defer s.wait.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateStaticMonitor()
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// updateMonitor 同步采集由 util 完成，service 负责发布缓存和记录采样历史。
func (s *service) updateMonitor() {
	snapshot := s.resourceSampler.Sample(s.ctx)
	network := make([]psutil.NetworkIO, 0, len(snapshot.Network))
	netReady, diskReady := false, false
	for _, value := range snapshot.Network {
		netReady = netReady || value.RateAvailable
		if !value.IsLoopback {
			network = append(network, value)
		}
	}
	snapshot.Network = network
	var netErr, diskErr error
	if message := snapshot.Errors["network"]; message != "" {
		netErr = errors.New(message)
	}
	if message := snapshot.Errors["disk"]; message != "" {
		diskErr = errors.New(message)
	}
	s.monitorMu.Lock()
	if netErr == nil {
		s.netRateCache = make(map[string]psutil.NetworkIO, len(network))
		for _, value := range network {
			s.netRateCache[value.Name] = value
		}
	}
	if diskErr == nil {
		s.diskRateCache = make(map[string]psutil.DiskIO, len(snapshot.Disk))
		for _, value := range snapshot.Disk {
			s.diskRateCache[value.Name] = value
			diskReady = diskReady || value.RateAvailable
		}
	}
	s.resourceCache = snapshot
	if netReady || diskReady {
		s.monitorUpdatedAt = snapshot.Time
		s.monitorSequence = max(s.monitorSequence+1, snapshot.Time)
		sample := monitorSample{
			ID:      s.monitorSequence,
			Time:    snapshot.Time,
			Network: s.netRateCache,
			Disk:    s.diskRateCache,
		}
		if len(s.monitorHistory) == 120 {
			copy(s.monitorHistory, s.monitorHistory[1:])
			s.monitorHistory[len(s.monitorHistory)-1] = sample
		} else {
			s.monitorHistory = append(s.monitorHistory, sample)
		}
	}
	s.monitorMu.Unlock()
	if netErr == nil && !netReady {
		netErr = errors.New("网卡计数器正在建立采样基线")
	}
	if diskErr == nil && !diskReady {
		diskErr = errors.New("磁盘计数器正在建立采样基线")
	}
	s.recordMonitor(snapshot, netErr, diskErr)
}

// updateStaticMonitor 低频刷新主机、磁盘容量与汇总设备，避免慢挂载点阻塞实时采样。
func (s *service) updateStaticMonitor() {
	systemBase, _ := psutil.Host(s.ctx)
	disks, diskErr := psutil.Disks(s.ctx)
	s.monitorMu.Lock()
	s.systemBaseCache = systemBase
	s.disksInfoCache = disks
	s.monitorMu.Unlock()
	s.updateMonitorDevices(diskErr == nil)
}
