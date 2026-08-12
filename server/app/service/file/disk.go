package file

import (
	"context"
	"github.com/shirou/gopsutil/v4/disk"
	"runtime"
	"server/app/util/file"
	"strings"
)

func (s *service) GetDisks() ([]Disk, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	excludedFs := map[string]struct{}{
		"tmpfs": {}, "devtmpfs": {}, "overlay": {},
		"squashfs": {}, "cdrom": {}, "vfat": {},
	}
	var disks []Disk
	for _, p := range partitions {
		// 基础校验
		if p.Fstype == "" || p.Device == "" || p.Mountpoint == "" {
			continue
		}
		// 过滤特殊文件系统
		if _, ok := excludedFs[strings.ToLower(p.Fstype)]; ok {
			continue
		}
		// 获取磁盘使用数据
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total <= 0 {
			continue
		}
		// 构建路径格式
		path := p.Mountpoint
		if runtime.GOOS == "windows" && len(path) >= 2 && path[1] == ':' {
			path = strings.ToUpper(string(path[0])) + "://"
		}
		// 构建inode数据
		inodeUsed, inodeFree := int64(usage.InodesUsed), int64(usage.InodesFree)
		// 计算百分比（避免除零）
		inodePercent := 0
		if usage.InodesTotal > 0 {
			inodePercent = int(float64(inodeUsed) / float64(usage.InodesTotal) * 100)
		}

		disks = append(disks, Disk{
			Filesystem: p.Device,
			Type:       p.Fstype,
			Path:       path,
			Size: struct {
				Total      int64 `json:"total"`
				Used       int64 `json:"used"`
				UnUsed     int64 `json:"un_used"`
				Percentage int   `json:"percentage"`
			}{
				Total:      int64(usage.Total),
				Used:       int64(usage.Used),
				UnUsed:     int64(usage.Free),
				Percentage: int(usage.UsedPercent),
			},
			Inodes: struct {
				Total      int64 `json:"total"`
				Used       int64 `json:"used"`
				UnUsed     int64 `json:"un_used"`
				Percentage int   `json:"percentage"`
			}{
				Total:      int64(usage.InodesTotal),
				Used:       inodeUsed,
				UnUsed:     inodeFree,
				Percentage: inodePercent,
			},
		})
	}
	return disks, nil
}

// GetDiskPaths 获取磁盘路径
func (s *service) GetDiskPaths() ([]string, error) {
	partitions, err := s.GetDisks()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, p := range partitions {
		paths = append(paths, p.Path)
	}
	return paths, nil
}

// CopyWithContext 带上下文控制的文件复制
// ctx: 上下文，用于取消操作
// src: 源路径
// destDir: 目标目录
// callback: 进度回调函数
func (s *service) CopyWithContext(ctx context.Context, src, destDir string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	f := file.NewDisk("")
	return f.CopyWithProcess(ctx, src, destDir, callback)
}

// MoveWithContext 带上下文控制的文件移动
// ctx: 上下文，用于取消操作
// src: 源路径
// destDir: 目标目录
// callback: 进度回调函数
func (s *service) MoveWithContext(ctx context.Context, src, destDir string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error {
	f := file.NewDisk("")
	return f.MoveWithProcess(ctx, src, destDir, callback)
}
