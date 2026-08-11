package script

import "server/app/model"

// FindById 查询常用脚本详情
func (s *service) FindById(id int64) (*model.Script, error) {
	return s.repositoryScript.FindById(id)
}
