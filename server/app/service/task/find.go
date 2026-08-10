package task

import (
	"server/app/model"
)

// findById 通过ID查询
func (s *service) findById(id int64) (*model.Task, error) {
	return s.repositoryTask.FindById(id)
}

// FindById 通过ID查询任务详情
func (s *service) FindById(id int64) (*Info, error) {
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
