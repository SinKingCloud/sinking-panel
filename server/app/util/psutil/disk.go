package psutil

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

// Disks 返回可浏览挂载卷，按路径排序，过滤系统虚拟文件系统和 macOS 隐藏卷。
// 单个挂载点不可读时跳过；所有候选挂载点均读取失败时返回错误。
func Disks(ctx context.Context) ([]DiskInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	partitions, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	result := make([]DiskInfo, 0, len(partitions))
	var failures []error
	for _, partition := range partitions {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if partition.Fstype == "" || partition.Device == "" || partition.Mountpoint == "" {
			continue
		}
		switch strings.ToLower(partition.Fstype) {
		case "autofs", "cgroup", "cgroup2", "configfs", "debugfs", "devfs", "devpts", "devtmpfs",
			"efivarfs", "fusectl", "hugetlbfs", "mqueue", "overlay", "proc", "procfs", "pstore", "ramfs",
			"securityfs", "squashfs", "sysfs", "tmpfs", "tracefs", "vfat":
			continue
		}
		// macOS 系统卷已通过根目录统一访问，避免重复展示隐藏挂载点。
		if runtime.GOOS == "darwin" && partition.Mountpoint != "/" &&
			(partition.Mountpoint == "/System/Volumes" || strings.HasPrefix(partition.Mountpoint, "/System/Volumes/") || slices.Contains(partition.Opts, "nobrowse")) {
			continue
		}
		usage, err := DiskUsage(ctx, partition.Mountpoint)
		if err != nil {
			failures = append(failures, fmt.Errorf("读取磁盘 %s 失败: %w", partition.Mountpoint, err))
			continue
		}
		if usage.Size.Total <= 0 {
			failures = append(failures, fmt.Errorf("磁盘 %s 未返回有效容量", partition.Mountpoint))
			continue
		}
		path := partition.Mountpoint
		if runtime.GOOS == "windows" && len(path) >= 2 && path[1] == ':' {
			path = strings.ToUpper(path[:1]) + "://"
		}
		result = append(result, DiskInfo{
			Filesystem: partition.Device,
			Type:       partition.Fstype,
			Path:       path,
			Size:       usage.Size,
			Inodes:     usage.Inodes,
		})
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 && len(failures) > 0 {
		return nil, errors.Join(failures...)
	}
	slices.SortFunc(result, func(a, b DiskInfo) int { return strings.Compare(a.Path, b.Path) })
	return result, nil
}

// DiskUsage 读取路径所在文件系统的容量；失败时返回错误，不把失败记为零用量。
func DiskUsage(ctx context.Context, path string) (DiskUsageInfo, error) {
	result := DiskUsageInfo{Path: path}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if strings.TrimSpace(path) == "" {
		return result, errors.New("磁盘路径不能为空")
	}
	usage, err := disk.UsageWithContext(ctx, filepath.Clean(path))
	if err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	// APFS 根卷与数据卷共享空间，不能只统计只读系统卷的已用量。
	if runtime.GOOS == "darwin" && filepath.Clean(path) == "/" && usage.Fstype == "apfs" && usage.Total > 0 && usage.Free <= usage.Total {
		usage.Used = usage.Total - usage.Free
		usage.UsedPercent = float64(usage.Used) / float64(usage.Total) * 100
	}
	result.Size = DiskSpaceInfo{
		Total:      int64(usage.Total),
		Used:       int64(usage.Used),
		UnUsed:     int64(usage.Free),
		Percentage: int(usage.UsedPercent),
	}
	result.Inodes = DiskSpaceInfo{
		Total:  int64(usage.InodesTotal),
		Used:   int64(usage.InodesUsed),
		UnUsed: int64(usage.InodesFree),
	}
	if usage.InodesTotal > 0 {
		result.Inodes.Percentage = int(float64(usage.InodesUsed) / float64(usage.InodesTotal) * 100)
	}
	return result, nil
}
