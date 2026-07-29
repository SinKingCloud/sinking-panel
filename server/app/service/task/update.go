package task

import (
	"errors"
	"server/app/enum/task_status"
	"server/app/enum/task_type"
	repositoryTask "server/app/repository/task"
	"time"
)

// updateEntryIDById 通过ID更新entryID
func (s *service) updateEntryIDById(id int64, entryID int) error {
	return s.repositoryTask.UpdateEntryIDById(id, entryID)
}

// updateStateById 通过ID更新实例和状态
func (s *service) updateStateById(id int64, entryID int, status int) error {
	return s.repositoryTask.UpdateStateById(id, entryID, status)
}

// updateRuntimeById 通过ID更新运行时间
func (s *service) updateRuntimeById(id int64, t time.Time) error {
	return s.repositoryTask.UpdateRuntimeById(id, t)
}

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryTask.UpdateTask) (err error) {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	if data.Type != nil {
		value, ok := data.Type.(int)
		if !ok {
			return errors.New("任务类型不合法")
		}
		if _, ok = task_type.Map()[value]; !ok {
			return errors.New("任务类型不合法")
		}
	}
	if data.Status != nil {
		value, ok := data.Status.(int)
		if !ok {
			return errors.New("任务状态不合法")
		}
		if _, ok = task_status.Map()[value]; !ok {
			return errors.New("任务状态不合法")
		}
	}
	if data.Name != nil {
		if value, ok := data.Name.(string); !ok || value == "" {
			return errors.New("任务名称不合法")
		}
	}
	if data.Script != nil {
		if value, ok := data.Script.(string); !ok || value == "" {
			return errors.New("任务内容不合法")
		}
	}
	if data.Spec != nil {
		value, ok := data.Spec.(string)
		if !ok || value == "" || !s.ValidateCron(value) {
			return errors.New("任务表达式不合法")
		}
	}
	err = s.repositoryTask.UpdateByIds(ids, data)
	if err == nil {
		for _, v := range ids {
			if refreshErr := s.Refresh(v); refreshErr != nil && err == nil {
				err = refreshErr
			}
		}
	}
	return
}
