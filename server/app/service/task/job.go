package task

import (
	"context"
	"server/app/enum/task_exec_type"
	"server/app/model"
	"server/app/util/str"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type job struct {
	*model.Task
	service *service
	ctx     context.Context
	run     RunFunc
	cancel  context.CancelFunc
	wait    sync.WaitGroup
	running int
	entryID cron.EntryID
	retired bool
}

// newJob 校验任务并固定本次执行内容。
func (s *service) newJob(task *model.Task) *job {
	if task == nil || s == nil || task.Spec == "" {
		return nil
	}
	content, err := s.checkContent(task.Script, task.ExecType)
	if err != nil {
		return nil
	}
	task.Script = content
	value := *task
	ctx, cancel := context.WithCancel(s.ctx)
	result := &job{
		Task:    &value,
		service: s,
		ctx:     ctx,
		cancel:  cancel,
	}
	s.jobs[result] = struct{}{}
	return result
}

// Run 统一跟踪调度器启动的任务，取消任务后旧的调度快照不再执行。
func (j *job) Run() {
	j.service.taskLock.Lock()
	value, ok := j.begin()
	j.service.taskLock.Unlock()
	if ok {
		j.execute(value)
	}
}

// begin 在 taskLock 内登记执行，销毁方先禁止新执行，再安全等待已有执行完成。
func (j *job) begin() (model.Task, bool) {
	if j.service.closed || j.retired || j.ctx.Err() != nil {
		return model.Task{}, false
	}
	// 内部任务复用同一实例，手动执行、改周期和恢复调度都不会重叠。
	if j.run != nil && j.running > 0 {
		return model.Task{}, false
	}
	j.running++
	j.wait.Add(1)
	j.service.runWait.Add(1)
	return *j.Task, true
}

// execute 运行固定内容，业务执行器须同步完成并响应上下文取消。
func (j *job) execute(task model.Task) {
	defer func() {
		j.service.taskLock.Lock()
		j.running--
		if j.running == 0 && j.entryID == 0 && task.Id >= 0 {
			j.retired = true
			j.cancel()
			delete(j.service.jobs, j)
		}
		j.service.taskLock.Unlock()
		j.wait.Done()
		j.service.runWait.Done()
	}()
	defer func() {
		if task.Id < 0 {
			j.service.taskLock.Lock()
			j.RunTime = str.DateTime(time.Now())
			j.service.taskLock.Unlock()
		} else {
			_ = j.service.repositoryTask.UpdateRuntimeById(task.Id, time.Now())
		}
	}()
	writeLog := func(content string) { _ = j.service.WriteLog(task.Id, content) }
	if j.run != nil {
		if err := j.run(j.ctx, task.Id, task.Script, writeLog); err != nil {
			writeLog(err.Error())
		}
		return
	}
	switch task.ExecType {
	case task_exec_type.Script:
		_ = j.service.script.Run(j.ctx, task.Script, writeLog)
	case task_exec_type.Request:
		if err := j.service.request.Run(j.ctx, task.Script, writeLog); err != nil {
			writeLog(err.Error())
		}
	}
}
