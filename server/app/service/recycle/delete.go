package recycle

import (
	"errors"
	"os"
	"path/filepath"
	"server/app/util/file"
)

// Delete 彻底删除文件
// name: 原始目录或文件路径
func (s *service) Delete(name string) error {
	s.lock.Lock()
	f := file.NewDisk(s.path)
	if name == "." || name == ".." || filepath.Base(name) != name {
		s.lock.Unlock()
		return errors.New("文件名称不合法")
	}
	if !f.Exists(name) {
		s.lock.Unlock()
		return errors.New("该目录或文件不存在")
	}
	tempPath, err := os.MkdirTemp(filepath.Dir(filepath.Clean(s.path)), ".recycle-delete-")
	if err != nil {
		s.lock.Unlock()
		return err
	}
	err = os.Rename(filepath.Join(filepath.Clean(s.path), name), filepath.Join(tempPath, "data"))
	s.lock.Unlock()
	if err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return os.RemoveAll(tempPath)
}

// Clear 清空回收站
// name: 原始目录或文件路径
func (s *service) Clear() error {
	s.lock.Lock()
	f := file.NewDisk("")
	if !f.Exists(s.path) {
		s.lock.Unlock()
		return nil
	}
	sourcePath := filepath.Clean(s.path)
	tempPath, err := os.MkdirTemp(filepath.Dir(sourcePath), ".recycle-delete-")
	if err != nil {
		s.lock.Unlock()
		return err
	}
	targetPath := filepath.Join(tempPath, "data")
	if err = os.Rename(sourcePath, targetPath); err != nil {
		s.lock.Unlock()
		_ = os.Remove(tempPath)
		return err
	}
	if err = os.MkdirAll(sourcePath, 0755); err != nil {
		if err2 := os.Rename(targetPath, sourcePath); err2 != nil {
			err = errors.Join(err, err2)
		}
		s.lock.Unlock()
		_ = os.Remove(tempPath)
		return err
	}
	s.lock.Unlock()
	return os.RemoveAll(tempPath)
}
