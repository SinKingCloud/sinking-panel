package task

import (
	"server/app/model"
	repositoryTask "server/app/repository/task"
	"server/app/util/page"
)

// SelectAll 获取所有数据
func (s *service) selectAll() ([]*model.Task, error) {
	return s.repositoryTask.SelectAll()
}

// Select 获取数据
func (s *service) Select(where *repositoryTask.SelectTask, queryPage *page.Query) (*page.Result[*repositoryTask.Task], error) {
	return s.repositoryTask.Select(where, queryPage)
}
