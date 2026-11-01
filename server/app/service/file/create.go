package file

import (
	"errors"
	"fmt"
	"server/app/util/file"
)

// Create 创建文件或目录，并设置指定权限。
func (s *service) Create(path, name string, chmod uint32) error {
	f := file.NewDisk(path)
	if f.Exists(name) {
		return errors.New("该目录或文件已存在")
	}
	if err := f.AutoCreate(name); err != nil {
		return errors.New("创建目录或文件失败")
	}
	if chmod > 0 {
		if err := f.SetFileMode(name, chmod); err != nil {
			return fmt.Errorf("修改权限失败: %w", err)
		}
	}
	return nil
}
