package recycle

import (
	"server/app/util/file"
)

// Count 统计数据
func (s *Service) Count() (totalSize int64, fileCount int64, dirCount int64, err error) {
	f := file.NewDisk(path)
	totalSize, fileCount, dirCount, err = f.Count("./")
	dirCount--
	return totalSize, fileCount, dirCount, err
}
