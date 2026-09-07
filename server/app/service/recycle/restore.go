package recycle

import (
	"errors"
	"path/filepath"
	"server/app/util/file"
)

// Restore 恢复文件
// name: 加密目录或文件路径
func (s *service) Restore(name string, path2 string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	f := file.NewDisk("")
	if name == "." || name == ".." || filepath.Base(name) != name {
		return errors.New("文件名称不合法")
	}
	oldPath := filepath.Join(s.path, name)
	if !f.Exists(oldPath) {
		return errors.New("该目录或文件不存在")
	}
	newPath, _, _ := s.decodeName(name)
	if newPath == "" {
		return errors.New("解码回收站文件失败")
	}
	if path2 == "" {
		if f.Move(oldPath, filepath.Dir(newPath)) == nil {
			if f.Rename(filepath.Join(filepath.Dir(newPath), filepath.Base(oldPath)), newPath) == nil {
				return nil
			}
			return errors.New("重命名回收站文件失败")
		}
		return errors.New("移动回收站文件失败")
	} else {
		path2 = f.Path(path2, true, true)
		_ = f.CreateDir(path2)
		if f.Move(oldPath, path2) == nil {
			if f.Rename(filepath.Join(path2, filepath.Base(oldPath)), filepath.Join(path2, filepath.Base(newPath))) == nil {
				return nil
			}
			return errors.New("重命名回收站文件失败")
		}
		return errors.New("移动回收站文件失败")
	}
}
