package task

import (
	"errors"
	"server/app/enum/task_exec_type"
	"server/app/enum/task_status"
	repositoryTask "server/app/repository/task"
	"server/app/util/page"
)

// Select 获取数据
func (s *service) Select(where *repositoryTask.SelectTask, queryPage *page.Query) (*page.Result[*repositoryTask.Task], error) {
	if where != nil && where.TypeId != nil {
		if *where.TypeId < 0 {
			return nil, errors.New("任务分类参数错误")
		}
	}
	if where != nil && where.ExecType != nil {
		if _, ok := task_exec_type.Map()[*where.ExecType]; !ok {
			return nil, errors.New("任务执行类型参数不合法")
		}
	}
	if where != nil && where.Status != nil {
		if _, ok := task_status.Map()[*where.Status]; !ok {
			return nil, errors.New("任务状态参数不合法")
		}
	}
	return s.repositoryTask.Select(where, queryPage)
}
