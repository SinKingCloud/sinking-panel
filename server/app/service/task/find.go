package task

import (
	"server/app/model"
)

// findById 通过ID查询
func (s *service) findById(id int64) (*model.Task, error) {
	return s.repositoryTask.FindById(id)
}

// FindById 通过ID查询
func (s *service) FindById(id int64) (data *model.Task, err error) {
	return s.findById(id)
}
