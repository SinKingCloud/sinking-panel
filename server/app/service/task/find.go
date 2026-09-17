package task

import (
	"errors"
	"server/app/model"

	"github.com/robfig/cron/v3"
)

// findById 在 taskLock 内查询，内部任务只从内存获取。
func (s *service) findById(id int64) (*model.Task, error) {
	if id < 0 {
		if task := s.internalTasks[id]; task != nil {
			return task.Task, nil
		}
		return nil, errors.New("任务不存在")
	}
	data, err := s.repositoryTask.FindById(id)
	if err == nil && data != nil && data.EntryID > 0 {
		// 调度编号仅在当前进程有效，不能用数据库中的旧编号操作其他任务。
		entry, ok := s.instance.Entry(cron.EntryID(data.EntryID)).Job.(*job)
		if !ok || entry.Id != id {
			data.EntryID = 0
		}
	}
	return data, err
}

// FindById 通过ID查询任务详情
func (s *service) FindById(id int64) (*Info, error) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	data, err := s.findById(id)
	if err != nil || data == nil {
		return nil, err
	}
	return &Info{
		Id:         data.Id,
		TypeId:     data.TypeId,
		ExecType:   data.ExecType,
		Name:       data.Name,
		Spec:       data.Spec,
		Script:     s.formatContent(data.Script, data.ExecType),
		Status:     data.Status,
		RunTime:    data.RunTime,
		CreateTime: data.CreateTime,
		UpdateTime: data.UpdateTime,
	}, nil
}
