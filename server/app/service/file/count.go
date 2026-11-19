package file

import (
	"errors"
	"fmt"
)

// Count 统计路径下的总大小、文件数和目录数。
func (s *service) Count(path string) (map[string]interface{}, error) {
	if !s.disk.Exists(path) {
		return nil, errors.New("该目录或文件不存在")
	}
	totalSize, fileCount, dirCount, err := s.disk.Count(path)
	if err != nil {
		return nil, fmt.Errorf("获取失败:%w", err)
	}
	return map[string]interface{}{"size": totalSize, "file": fileCount, "dir": dirCount}, nil
}
