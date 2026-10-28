package database

import (
	"errors"
	"server/app/constant"
	"server/app/enum/script_module"
	"strconv"

	"gorm.io/gorm"
)

// DeleteByIds 只删除面板保存的实例连接，不操作远端数据库。
func (s *service) DeleteByIds(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("请选择实例")
	}
	for _, id := range ids {
		if id <= 0 {
			return errors.New("实例 ID 不合法")
		}
	}
	modules := make([]string, 0, len(ids))
	for _, id := range ids {
		modules = append(modules, script_module.Database+"_"+strconv.FormatInt(id, 10))
	}
	if err := s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repositoryScript.DeleteByModules(modules, tx); err != nil {
			return err
		}
		if err := s.repositoryTypes.DeleteByModules(modules, tx); err != nil {
			return err
		}
		return s.repository.DeleteByIds(ids, tx)
	}); err != nil {
		return err
	}
	for _, module := range modules {
		s.cache.Delete(constant.CacheNameWithTypeEnum + module)
	}
	return nil
}
