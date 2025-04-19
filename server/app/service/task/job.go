package task

import (
	"server/app/constant"
	"server/app/model"
	"server/app/util/cmd"
	"time"
)

type job struct {
	*model.Task
}

func newJob(task *model.Task) *job {
	if task.Spec == "" || task.Script == "" {
		return nil
	}
	return &job{task}
}

func (j *job) Run() {
	defer func() {
		_ = obj.updateRuntimeById(j.Id, time.Now())
	}()
	switch j.Type {
	case int(Script):
		c := cmd.NewScriptExec(constant.TempPath, 43200, func(s string) {
			_ = obj.WriteLog(j.Id, s)
		})
		_, _, _ = c.Execute(j.Script)
		break
	}
}
