package system

import (
	"fmt"
	"server/app/enum/system_task_status"
	"time"
)

// taskWorker 从共享队列取任务。worker 数量固定，任务数量不会直接创建同等数量的 goroutine。
func (s *service) taskWorker() {
	defer s.wait.Done()
	for {
		job, ok := s.nextTask()
		if !ok {
			return
		}
		s.runTask(job)
	}
}

func (s *service) nextTask() (*job, bool) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	for {
		for len(s.queue) == 0 && s.ctx.Err() == nil {
			s.queueCond.Wait()
		}
		if s.ctx.Err() != nil {
			return nil, false
		}
		id := s.queue[0]
		s.queue[0] = ""
		s.queue = s.queue[1:]
		job := s.jobs[id]
		delete(s.jobs, id)
		if job != nil {
			return job, true
		}
	}
}

func (s *service) runTask(job *job) {
	if job == nil {
		return
	}

	s.mu.Lock()
	task, ok := s.tasks[job.id]
	if !ok || task.Status != system_task_status.Pending || job.ctx.Err() != nil {
		s.mu.Unlock()
		return
	}
	task.Status = system_task_status.Running
	task.StartTime = time.Now().Unix()
	task.UpdateTime = task.StartTime
	s.mu.Unlock()
	s.appendTaskLog(job.id, system_task_status.Running, 0, "任务开始执行")

	update := func(status int, progress float64, message string) {
		s.TaskUpdate(job.id, status, progress, message)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			s.TaskUpdate(job.id, system_task_status.Failed, 0, "任务执行异常: "+fmt.Sprint(recovered))
		}
		s.finishTask(job)
	}()

	if job.run == nil {
		s.TaskUpdate(job.id, system_task_status.Failed, 0, "任务执行器为空")
		return
	}
	job.run(job.ctx, job.data, update)
}

func (s *service) finishTask(job *job) {
	s.mu.RLock()
	task, ok := s.tasks[job.id]
	status := system_task_status.Failed
	if ok {
		status = task.Status
	}
	s.mu.RUnlock()
	terminal := status == system_task_status.Completed || status == system_task_status.Failed || status == system_task_status.Canceled
	if !ok || terminal {
		return
	}
	if job.ctx.Err() != nil {
		s.TaskUpdate(job.id, system_task_status.Canceled, 0, "任务已取消")
		return
	}
	s.TaskUpdate(job.id, system_task_status.Failed, 0, "任务未正常结束")
}
