package types

import (
	"errors"
	"server/app/model"
	"server/app/util/str"
	"strings"
	"unicode/utf8"
)

// Create 插入数据
func (s *service) Create(data *model.Type) error {
	if data == nil {
		return errors.New("类型数据不能为空")
	}
	if strings.TrimSpace(data.Module) == "" || utf8.RuneCountInString(data.Module) > 50 || strings.Contains(data.Module, "%") {
		return errors.New("所属模块不能为空、超过50个字符或包含通配符")
	}
	data.Id = str.GetSnowWorkIns().GetId()
	data.Sort = data.Id
	if err := s.repositoryTypes.Create(data); err != nil {
		return err
	}
	s.clearTypeEnumCache(data.Module)
	return nil
}
