package file

import (
	"context"
	"server/app/util/file"
	"server/app/util/psutil"
)

// GetDisks 返回文件管理可浏览的磁盘，采集由 psutil 统一处理。
func (s *service) GetDisks() ([]Disk, error) {
	values, err := psutil.Disks(context.Background())
	if err != nil {
		return nil, err
	}
	result := make([]Disk, 0, len(values))
	for _, value := range values {
		result = append(result, Disk{
			Filesystem: value.Filesystem,
			Type:       value.Type,
			Path:       value.Path,
			Size:       value.Size,
			Inodes:     value.Inodes,
		})
	}
	return result, nil
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
