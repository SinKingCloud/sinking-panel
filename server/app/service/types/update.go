package types

import (
	"errors"
	"server/app/enum/type_module"
	repositoryTypes "server/app/repository/types"

	"gorm.io/gorm"
)

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryTypes.UpdateType) error {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	var module string
	if data.Module != nil {
		value, ok := data.Module.(string)
		if !ok {
			return errors.New("所属模块不合法")
		}
		if _, ok = type_module.Map()[value]; !ok {
			return errors.New("所属模块不合法")
		}
		module = value
	}
	modules := make([]string, 0)
	err := s.database.Transaction(func(tx *gorm.DB) error {
		types, err := s.repositoryTypes.SelectByIds(ids, tx)
		if err != nil {
			return err
		}
		scriptTypeIds := make([]int64, 0)
		siteTypeIds := make([]int64, 0)
		taskTypeIds := make([]int64, 0)
		for _, item := range types {
			modules = append(modules, item.Module)
			if item.Module == type_module.Script {
				scriptTypeIds = append(scriptTypeIds, item.Id)
			}
			if item.Module == type_module.Site {
				siteTypeIds = append(siteTypeIds, item.Id)
			}
			if item.Module == type_module.Task {
				taskTypeIds = append(taskTypeIds, item.Id)
			}
		}
		if module != "" && module != type_module.Script {
			if err := s.repositoryScript.ClearTypeId(scriptTypeIds, tx); err != nil {
				return err
			}
		}
		if module != "" && module != type_module.Site {
			if err := s.repositorySite.ClearTypeId(siteTypeIds, tx); err != nil {
				return err
			}
		}
		if module != "" && module != type_module.Task {
			if err := s.repositoryTask.ClearTypeId(taskTypeIds, tx); err != nil {
				return err
			}
		}
		return s.repositoryTypes.UpdateByIds(ids, data, tx)
	})
	if err != nil {
		return err
	}
	if module != "" {
		modules = append(modules, module)
	}
	s.clearTypeEnumCache(modules...)
	return nil
}
