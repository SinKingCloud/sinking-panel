package types

import (
	"errors"
	"server/app/enum/type_module"
	repositoryTypes "server/app/repository/types"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryTypes.UpdateType) error {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	reorder := data.Reorder != nil && *data.Reorder
	if reorder && (data.Module != nil || data.Name != nil || data.Sort != nil) {
		return errors.New("调整分类顺序时不能同时修改其他信息")
	}
	var module string
	if data.Module != nil {
		if strings.TrimSpace(*data.Module) == "" || utf8.RuneCountInString(*data.Module) > 50 || strings.Contains(*data.Module, "%") {
			return errors.New("所属模块不能为空、超过50个字符或包含通配符")
		}
		module = *data.Module
	}
	modules := make([]string, 0)
	err := s.database.Transaction(func(tx *gorm.DB) error {
		types, err := s.repositoryTypes.SelectByIds(ids, tx)
		if err != nil {
			return err
		}
		if reorder && len(types) != len(ids) {
			return errors.New("分类列表已变更，请刷新后重试")
		}
		scriptTypeIds := make([]int64, 0)
		siteTypeIds := make([]int64, 0)
		taskTypeIds := make([]int64, 0)
		databaseTypeIds := make([]int64, 0)
		// 改名、排序只更新分类；变更模块时才清理原业务关联。
		for _, item := range types {
			modules = append(modules, item.Module)
			if module == "" || module == item.Module {
				continue
			}
			scriptTypeIds = append(scriptTypeIds, item.Id)
			switch item.Module {
			case type_module.Site:
				siteTypeIds = append(siteTypeIds, item.Id)
			case type_module.Task:
				taskTypeIds = append(taskTypeIds, item.Id)
			case type_module.Database:
				databaseTypeIds = append(databaseTypeIds, item.Id)
			}
		}
		if err := s.repositoryScript.ClearTypeId(scriptTypeIds, tx); err != nil {
			return err
		}
		if err := s.repositorySite.ClearTypeId(siteTypeIds, tx); err != nil {
			return err
		}
		if err := s.repositoryTask.ClearTypeId(taskTypeIds, tx); err != nil {
			return err
		}
		if err := s.repositoryDatabase.ClearTypeId(databaseTypeIds, tx); err != nil {
			return err
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
