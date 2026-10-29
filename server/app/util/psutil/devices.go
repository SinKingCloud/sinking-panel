package psutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

var darwinPartition = regexp.MustCompile(`^(disk[0-9]+)s[0-9]+`)

// AggregateDevices 选择适合汇总整机 I/O 的设备，避免分区、网桥与底层设备重复计数。
// 这是启发式选择；明细查询仍保留所有设备。输入不修改，输出按名称排序。
func AggregateDevices(diskNames, networkNames []string) (disks, networks []string) {
	diskSet := make(map[string]bool, len(diskNames))
	for _, name := range diskNames {
		diskSet[name] = true
	}
	disks = make([]string, 0, len(diskNames))
	for _, name := range diskNames {
		if aggregateDiskDevice(name, diskSet, runtime.GOOS, "/sys/class/block") {
			disks = append(disks, name)
		}
	}
	physical := make([]string, 0)
	networks = make([]string, 0, len(networkNames))
	for _, name := range networkNames {
		if !aggregateNetworkDevice(name) {
			continue
		}
		networks = append(networks, name)
		if runtime.GOOS == "linux" {
			if _, err := os.Stat(filepath.Join("/sys/class/net", name, "device")); err == nil {
				physical = append(physical, name)
			}
		}
	}
	if len(physical) > 0 {
		networks = physical
	}
	slices.Sort(disks)
	slices.Sort(networks)
	return disks, networks
}

func aggregateDiskDevice(name string, disks map[string]bool, platform, sysRoot string) bool {
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
		if match := darwinPartition.FindStringSubmatch(name); len(match) == 2 && disks[match[1]] {
			return false
		}
	}
	return true
}

func aggregateNetworkDevice(name string) bool {
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
