package script

import (
	"errors"
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/util/str"
)

// Create 创建常用脚本
func (s *service) Create(data *model.Script) error {
	if data == nil {
		return errors.New("常用脚本数据不能为空")
	}
	if data.TypeId < 0 {
		return errors.New("脚本分类不合法")
	}
	if data.TypeId > 0 {
		category, err := s.typeService.FindById(data.TypeId)
		if err != nil || category == nil || category.Module != type_module.Script {
			return errors.New("脚本分类不合法")
		}
	}
	data.Id = str.GetSnowWorkIns().GetId()
	return s.repositoryScript.Create(data)
}
