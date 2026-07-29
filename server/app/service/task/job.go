package task

import (
	"server/app/constant"
	"server/app/enum/task_type"
	"server/app/model"
	"server/app/util/cmd"
	"time"
)

type job struct {
	*model.Task
	service *service
}

func newJob(task *model.Task, service *service) *job {
	if task.Spec == "" || task.Script == "" {
		return nil
	}
	return &job{Task: task, service: service}
}

func (j *job) Run() {
	defer func() {
		_ = j.service.updateRuntimeById(j.Id, time.Now())
	}()
	switch j.Type {
	case task_type.Script:
		c := cmd.NewScriptExec(constant.TempPath, 43200, func(s string) {
			_ = j.service.WriteLog(j.Id, s)
		})
		_, _, _ = c.Execute(j.Script)
		break
	}
}
