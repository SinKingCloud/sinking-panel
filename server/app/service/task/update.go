package task

import (
	"errors"
	"server/app/enum/task_exec_type"
	"server/app/enum/task_status"
	repositoryTask "server/app/repository/task"
	"strings"
)

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryTask.UpdateTask) (err error) {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	if data.TypeId != nil {
		if err = s.validateTypeId(*data.TypeId); err != nil {
			return err
		}
	}
	if data.ExecType != nil {
		if _, ok := task_exec_type.Map()[*data.ExecType]; !ok {
			return errors.New("任务执行类型不合法")
		}
	}
	if data.Status != nil {
		if _, ok := task_status.Map()[*data.Status]; !ok {
			return errors.New("任务状态不合法")
		}
	}
	if data.Name != nil {
		if strings.TrimSpace(*data.Name) == "" {
			return errors.New("任务名称不合法")
		}
	}
	if data.Script != nil {
		if strings.TrimSpace(*data.Script) == "" {
			return errors.New("任务内容不合法")
		}
	}
	if data.Spec != nil {
		if *data.Spec == "" || !s.validateCron(*data.Spec) {
			return errors.New("任务表达式不合法")
		}
	}
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	if s.closed {
		return errors.New("计划任务服务已关闭")
	}
	var normalizedContent string
	contentType := -1
	if data.ExecType != nil || data.Script != nil {
		for _, id := range ids {
			task, findErr := s.repositoryTask.FindById(id)
			if findErr != nil {
				return findErr
			}
			execType := task.ExecType
			content := task.Script
			if data.ExecType != nil {
				execType = *data.ExecType
				if execType != task.ExecType && data.Script == nil {
					return errors.New("修改任务执行类型时必须同时提交任务内容")
				}
			}
			if data.Script != nil {
				content = *data.Script
				if data.ExecType == nil {
					if contentType >= 0 && contentType != execType {
						return errors.New("不同执行类型的任务不能批量修改内容")
					}
					contentType = execType
				}
			}
			value, checkErr := s.checkContent(content, execType)
			if checkErr != nil {
				return checkErr
			}
			if data.Script != nil {
				normalizedContent = value
			}
		}
	}
	if data.Script != nil {
		data.Script = &normalizedContent
	}
	err = s.repositoryTask.UpdateByIds(ids, data)
	if err == nil {
		for _, v := range ids {
			if refreshErr := s.refresh(v); refreshErr != nil && err == nil {
				err = refreshErr
			}
		}
	}
	return
}
