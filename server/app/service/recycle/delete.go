package recycle

import (
	"errors"
	"server/app/util/file"
)

// Delete 彻底删除文件
// name: 原始目录或文件路径
func (s *Service) Delete(name string) error {
	f := file.NewDisk(path)
	if !f.Exists(name) {
		return errors.New("该目录或文件不存在")
	}
	return f.Delete(name)
}

// Clear 清空回收站
// name: 原始目录或文件路径
func (s *Service) Clear() error {
	f := file.NewDisk("")
	if f.Exists(path) {
		return f.Delete(path)
	}
	return nil
}
