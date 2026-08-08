package task

import (
	"errors"
	"server/app/enum/task_status"
	"server/app/enum/task_type"
	"server/app/model"
	repositoryTask "server/app/repository/task"
	"server/app/util/page"
	"strconv"
)

// SelectAll 获取所有数据
func (s *service) selectAll() ([]*model.Task, error) {
	return s.repositoryTask.SelectAll()
}

// Select 获取数据
func (s *service) Select(where *repositoryTask.SelectTask, queryPage *page.Query) (*page.Result[*repositoryTask.Task], error) {
	if where != nil && where.Type != "" {
		value, err := strconv.Atoi(where.Type)
		if err != nil {
			return nil, errors.New("任务类型参数错误")
		}
		if _, ok := task_type.Map()[value]; !ok {
			return nil, errors.New("任务类型参数不合法")
		}
	}
	if where != nil && where.Status != "" {
		value, err := strconv.Atoi(where.Status)
		if err != nil {
			return nil, errors.New("任务状态参数错误")
		}
		if _, ok := task_status.Map()[value]; !ok {
			return nil, errors.New("任务状态参数不合法")
		}
	}
	return s.repositoryTask.Select(where, queryPage)
}
