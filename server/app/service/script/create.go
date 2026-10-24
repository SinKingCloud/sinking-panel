package script

import (
	"errors"
	"server/app/model"
	"server/app/util/str"
	"strings"
	"unicode/utf8"
)

// Create 创建常用脚本。
func (s *service) Create(data *model.Script) error {
	if data == nil {
		return errors.New("常用脚本数据不能为空")
	}
	if strings.TrimSpace(data.Module) == "" || utf8.RuneCountInString(data.Module) > 50 {
		return errors.New("脚本所属模块不能为空且不能超过50个字符")
	}
	if strings.TrimSpace(data.Script) == "" {
		return errors.New("脚本内容不能为空")
	}
	if strings.TrimSpace(data.Name) == "" {
		return errors.New("脚本名称不能为空")
	}
	if data.TypeId < 0 {
		return errors.New("脚本分类不合法")
	}
	if data.TypeId > 0 {
		categories, err := s.repositoryTypes.SelectByIds([]int64{data.TypeId})
		if err != nil {
			return err
		}
		module, _, _ := strings.Cut(data.Module, "_")
		if len(categories) != 1 || (categories[0].Module != data.Module && categories[0].Module != module) {
			return errors.New("脚本分类不属于当前保存范围")
		}
	}
	data.Id = str.GetSnowWorkIns().GetId()
	return s.repositoryScript.Create(data)
}
