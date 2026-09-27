package system

import (
	"errors"
	"server/app/enum/system_task_status"
)

// TaskList 获取任务列表
func (s *service) TaskList(status *int) ([]*Task, error) {
	if status != nil {
		if _, ok := system_task_status.Map()[*status]; !ok {
			return nil, errors.New("任务状态参数不合法")
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if status != nil && task.Status != *status {
			continue
		}
		snapshot := *task
		tasks = append(tasks, &snapshot)
	}
	return tasks, nil
}
