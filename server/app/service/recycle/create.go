package recycle

import (
	"errors"
	"path/filepath"
	"server/app/util/file"
)

// Create 添加回收站文件
// name: 文件或目录路径
func (s *Service) Create(name string) error {
	f := file.NewDisk("")
	if !f.Exists(name) {
		return errors.New("该目录或文件不存在")
	}
	newName := s.encodeName(f.Path(name, true, true), 0, 0)
	if f.Rename(name, filepath.Join(path, newName)) != nil {
		return errors.New("移动文件失败")
	}
	return nil
}
