package recycle

import (
	"errors"
	"os"
	"server/app/util/file"
)

// Count 统计数据
func (s *service) Count() (totalSize int64, fileCount int64, dirCount int64, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	f := file.NewDisk(path)
	if !f.Exists("./") {
		return 0, 0, 0, nil
	}
	totalSize, fileCount, dirCount, err = f.Count("./")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, 0, 0, nil
		}
		return 0, 0, 0, err
	}
	dirCount--
	return totalSize, fileCount, dirCount, err
}
