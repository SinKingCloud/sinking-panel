package recycle

import (
	"path/filepath"
	"server/app/util/file"
	"server/app/util/str"
	"time"
)

// File 描述文件系统对象的元数据信息
type File struct {
	ID         string       `json:"id"`          // 加密路径
	Name       string       `json:"name"`        // 原始名称
	Path       string       `json:"path"`        // 原始路径
	Size       int64        `json:"size"`        // 文件大小（字节），目录为0
	DeleteTime str.DateTime `json:"update_time"` //删除时间
}

// Select 获取数据
func (s *Service) Select(page int, pageSize int, orderByField string, orderByType string) (list []*File, total int64, err error) {
	f := file.NewDisk(path)
	var l []*file.File
	l, total, err = f.FileListWithPage("./", page, pageSize, orderByField, orderByType)
	if err != nil {
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
			DeleteTime: str.DateTime(time.Unix(deleteTime, 0)),
		})
	}
	return list, total, err
}
