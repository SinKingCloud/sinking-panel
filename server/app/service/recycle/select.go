package recycle

import (
	"errors"
	"os"
	"path/filepath"
	"server/app/util/file"
	"server/app/util/str"
	"time"
)

// Select 获取数据
func (s *service) Select(page int, pageSize int, orderByField string, orderByType string) (list []*File, total int64, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	f := file.NewDisk(s.path)
	if !f.Exists("./") {
		return []*File{}, 0, nil
	}
	var l []*file.File
	l, total, err = f.FileListWithPage("./", page, pageSize, orderByField, orderByType)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*File{}, 0, nil
		}
		return nil, 0, err
	}
	list = []*File{}
	for _, v := range l {
		old, deleteTime, _ := s.decodeName(v.Name)
		list = append(list, &File{
			ID:         v.Name,
			Name:       filepath.Base(old),
			Path:       old,
			Size:       v.Size,
			IsDir:      v.IsDir,
			DeleteTime: str.DateTime(time.Unix(deleteTime, 0)),
		})
	}
	return list, total, err
}
