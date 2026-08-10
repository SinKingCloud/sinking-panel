package types

import (
	"server/app/enum/type_module"

	"gorm.io/gorm"
)

// DeleteByIds 通过ID列表删除
func (s *service) DeleteByIds(ids []int64) error {
	modules := make([]string, 0)
	err := s.database.Transaction(func(tx *gorm.DB) error {
		types, err := s.repositoryTypes.SelectByIds(ids, tx)
		if err != nil {
			return err
		}
		scriptTypeIds := make([]int64, 0)
		for _, item := range types {
			modules = append(modules, item.Module)
			if item.Module == type_module.Script {
				scriptTypeIds = append(scriptTypeIds, item.Id)
			}
		}
		if err := s.repositoryScript.ClearTypeId(scriptTypeIds, tx); err != nil {
			return err
		}
		return s.repositoryTypes.DeleteByIds(ids, tx)
	})
	if err != nil {
		return err
	}
	s.clearTypeEnumCache(modules...)
	return nil
}
