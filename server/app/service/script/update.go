package script

import (
	"errors"
	"server/app/enum/type_module"
	repositoryScript "server/app/repository/script"
)

// UpdateByIds 通过ID列表更新常用脚本
func (s *service) UpdateByIds(ids []int64, data *repositoryScript.UpdateScript) error {
	if len(ids) == 0 || data == nil {
		return nil
	}
	if data.TypeId != nil {
		if *data.TypeId < 0 {
			return errors.New("脚本分类不合法")
		}
		if *data.TypeId > 0 {
			category, err := s.typeService.FindById(*data.TypeId)
			if err != nil || category == nil || category.Module != type_module.Script {
				return errors.New("脚本分类不合法")
			}
		}
	}
	return s.repositoryScript.UpdateByIds(ids, data)
}
