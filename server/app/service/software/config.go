package software

import (
	"errors"
	repositorySoftware "server/app/repository/software"
)

// GetConfig 按软件标识读取插件内部配置，内容格式由插件决定。
func (s *service) GetConfig(code string) (string, error) {
	record, err := s.record(code)
	if err != nil {
		return "", err
	}
	if record == nil {
		return "", errors.New("软件不存在")
	}
	return record.Config, nil
}

// SetConfig 按软件标识保存插件内部配置，支持清空；生命周期内也可调用。
func (s *service) SetConfig(code, config string) error {
	// 生命周期持有软件操作锁，配置读写不再获取该锁，避免插件回调重入死锁。
	record, err := s.record(code)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("软件不存在")
	}
	err = s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{Config: &config})
	s.invalidate(code)
	return err
}
