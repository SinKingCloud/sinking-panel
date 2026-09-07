package recycle

import (
	"path/filepath"
	"server/app/constant"
	"sync"
)

// Service service接口
type Service interface {
	Count() (totalSize int64, fileCount int64, dirCount int64, err error)
	Create(name string) error
	Delete(name string) error
	Clear() error
	Restore(name string, path string) error
	Select(page int, pageSize int, orderByField string, orderByType string) (list []*File, total int64, err error)
}

// service 注入结构
type service struct {
	lock sync.RWMutex
	path string // 回收站目录
}

// NewService 实例化service
func NewService() *service {
	return &service{path: filepath.Clean(constant.RecyclePath)}
}
