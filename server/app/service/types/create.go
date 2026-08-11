package types

import (
	"errors"
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/util/str"
)

// Create 插入数据
func (s *service) Create(data *model.Type) error {
	if data == nil {
		return errors.New("类型数据不能为空")
	}
	if _, ok := type_module.Map()[data.Module]; !ok {
		return errors.New("所属模块不合法")
	}
	data.Id = str.GetSnowWorkIns().GetId()
	data.Sort = data.Id
	if err := s.repositoryTypes.Create(data); err != nil {
		return err
	}
	s.clearTypeEnumCache(data.Module)
	return nil
}
