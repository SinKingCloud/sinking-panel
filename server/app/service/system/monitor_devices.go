package system

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	serviceFile "server/app/service/file"
	"strings"
)

var monitorDarwinPartition = regexp.MustCompile(`^(disk[0-9]+)s[0-9]+`)

// updateMonitorDevices 低频刷新汇总设备，避免每秒遍历 sysfs。
func (s *service) updateMonitorDevices(diskReady bool) {
	s.ioCacheLock.RLock()
	disks := make(map[string]bool, len(s.diskRateCache))
	for name := range s.diskRateCache {
		disks[name] = true
	}
	networks := make(map[string]bool, len(s.netRateCache))
	for name := range s.netRateCache {
		networks[name] = true
	}
	s.ioCacheLock.RUnlock()

	for name := range disks {
		if !monitorDiskDevice(name, disks, runtime.GOOS, "/sys/class/block") {
			delete(disks, name)
		}
	}
	physical := make(map[string]bool)
	for name := range networks {
		if !monitorNetworkDevice(name) {
			delete(networks, name)
			continue
		}
		if runtime.GOOS == "linux" {
			if _, err := os.Stat(filepath.Join("/sys/class/net", name, "device")); err == nil {
				physical[name] = true
			}
		}
	}
	// 物理接口已有完整流量时，不再累计网桥、VLAN、隧道等上层接口。
	// 容器中的 eth0 通常没有 device 链接，此时保留过滤后的接口。
	if len(physical) > 0 {
		networks = physical
	}
	s.monitorDeviceMu.Lock()
	s.monitorDisks = disks
	s.monitorNetworks = networks
	s.monitorDiskReady = diskReady
	s.monitorDeviceMu.Unlock()
}

func monitorDiskDevice(name string, disks map[string]bool, platform, sysRoot string) bool {
	name = filepath.Base(name)
	if platform == "linux" {
		for _, prefix := range []string{"loop", "ram", "zram", "fd", "sr"} {
			if strings.HasPrefix(name, prefix) {
				return false
			}
		}
		path := filepath.Join(sysRoot, name)
		if _, err := os.Stat(filepath.Join(path, "partition")); err == nil {
			return false
		}
		// dm/LVM、软 RAID 等叠加设备与底层磁盘共享 I/O，只计底层设备。
		if slaves, err := os.ReadDir(filepath.Join(path, "slaves")); err == nil && len(slaves) > 0 {
			return false
		}
	} else if platform == "darwin" {
		if match := monitorDarwinPartition.FindStringSubmatch(name); len(match) == 2 && disks[match[1]] {
			return false
		}
	}
	return true
}

func monitorNetworkDevice(name string) bool {
	lower := strings.ToLower(name)
	if lower == "lo" || lower == "lo0" || strings.Contains(lower, "loopback") {
		return false
	}
	for _, prefix := range []string{"veth", "docker", "br-", "virbr", "cni", "flannel", "kube", "tun", "tap", "utun", "bond", "team", "dummy", "ifb", "awdl", "llw", "bridge"} {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}
	return true
}

// monitorDiskCapacity 同一文件系统的多次挂载只计算一次。
func monitorDiskCapacity(disks []serviceFile.Disk) (used, total float64) {
	seen := make(map[string]bool, len(disks))
	for _, disk := range disks {
		key := disk.Filesystem
		if key == "" {
			key = disk.Path
		}
		if seen[key] || disk.Size.Total <= 0 {
			continue
		}
		seen[key] = true
		total += float64(disk.Size.Total)
		used += float64(max(0, min(disk.Size.Used, disk.Size.Total)))
	}
	return used, total
}
