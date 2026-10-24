package server

import (
	"server/app/constant"
	"server/app/enum/type_module"
	"strconv"

	"gorm.io/gorm"
)

// DeleteByIds 同一事务内删除终端及其专属脚本和分类。
func (s *service) DeleteByIds(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	modules := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			modules = append(modules, type_module.Ssh+"_"+strconv.FormatInt(id, 10))
		}
	}
	if err := s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repositoryScript.DeleteByModules(modules, tx); err != nil {
			return err
		}
		if err := s.repositoryTypes.DeleteByModules(modules, tx); err != nil {
			return err
		}
		return s.repositoryServer.DeleteByIds(ids, tx)
	}); err != nil {
		return err
	}
	for _, module := range modules {
		s.cache.Delete(constant.CacheNameWithTypeEnum + module)
	}
	return nil
}
