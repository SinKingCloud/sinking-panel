package file

import (
	"errors"
	"server/app/util/file"
)

// Delete 批量删除文件或目录，可选择移入回收站。
func (s *service) Delete(paths []string, recycle bool) error {
	if len(paths) == 0 {
		return errors.New("文件路径列表不能为空")
	}
	if recycle && s.recycleService == nil {
		return errors.New("回收站服务不可用")
	}
	f := file.NewDisk("")
	for _, path := range paths {
		if path == "" || !f.Exists(path) {
			return errors.New("该目录或文件不存在: " + path)
		}
	}
	for _, path := range paths {
		if recycle {
			if err := s.recycleService.Create(path); err != nil {
				return errors.New("移动至回收站失败")
			}
		} else if err := f.Delete(path); err != nil {
			return errors.New("删除失败")
		}
	}
	return nil
}
