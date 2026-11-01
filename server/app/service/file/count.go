package file

import (
	"errors"
	"fmt"
	"server/app/util/file"
)

// Count 统计路径下的总大小、文件数和目录数。
func (s *service) Count(path string) (map[string]interface{}, error) {
	f := file.NewDisk("")
	if !f.Exists(path) {
		return nil, errors.New("该目录或文件不存在")
	}
	totalSize, fileCount, dirCount, err := f.Count(path)
	if err != nil {
		return nil, fmt.Errorf("获取失败:%w", err)
	}
	return map[string]interface{}{"size": totalSize, "file": fileCount, "dir": dirCount}, nil
}
