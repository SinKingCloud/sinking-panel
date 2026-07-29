package task

import "server/app/model"

// create 插入数据
func (s *service) create(data *model.Task) error {
	return s.repositoryTask.Create(data)
}
