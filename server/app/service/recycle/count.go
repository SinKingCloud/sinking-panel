package recycle

import (
	"server/app/util/file"
)

// Count 统计数据
func (s *service) Count() (totalSize int64, fileCount int64, dirCount int64, err error) {
	f := file.NewDisk(path)
	if !f.Exists("./") {
		return 0, 0, 0, nil
	}
	totalSize, fileCount, dirCount, err = f.Count("./")
	dirCount--
	return totalSize, fileCount, dirCount, err
}
