package script

import (
	"errors"
	"server/app/model"
	"server/app/util/str"
)

// Create 创建常用脚本
func (s *service) Create(data *model.Script) error {
	if data == nil {
		return errors.New("常用脚本数据不能为空")
	}
	data.Id = str.GetSnowWorkIns().GetId()
	return s.repositoryScript.Create(data)
}
