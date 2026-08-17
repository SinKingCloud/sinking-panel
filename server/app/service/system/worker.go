package system

import (
	"server/app/enum/system_task_status"
	"time"
)

// taskWorker 从共享队列取任务。worker 数量固定，任务数量不会直接创建同等数量的 goroutine。
func (s *service) taskWorker() {
	for {
		job := s.nextTask()
		if job == nil {
			continue
		}
		s.runTask(job)
	}
}

func (s *service) nextTask() *job {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	for len(s.queue) == 0 {
		s.queueCond.Wait()
	}
	id := s.queue[0]
	s.queue[0] = ""
	s.queue = s.queue[1:]
	job := s.jobs[id]
	delete(s.jobs, id)
	return job
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

	update := func(status int, progress float64, message string) {
		s.TaskUpdate(job.id, status, progress, message)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			s.TaskUpdate(job.id, system_task_status.Failed, 0, "任务执行异常")
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

// taskCleanupWorker 统一清理终态任务，避免每个任务完成后额外启动延迟 goroutine。
func (s *service) taskCleanupWorker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		deadline := now.Unix() - int64(taskRetention/time.Second)
		ids := make([]string, 0)
		s.mu.RLock()
		for id, task := range s.tasks {
			terminal := task.Status == system_task_status.Completed || task.Status == system_task_status.Failed || task.Status == system_task_status.Canceled
			if terminal && task.EndTime > 0 && task.EndTime <= deadline {
				ids = append(ids, id)
			}
		}
		s.mu.RUnlock()
		for _, id := range ids {
			s.mu.Lock()
			task, ok := s.tasks[id]
			if !ok {
				s.mu.Unlock()
				continue
			}
			cancel := task.cancel
			delete(s.tasks, id)
			s.mu.Unlock()
			if cancel != nil {
				cancel()
			}
			s.queueMu.Lock()
			delete(s.jobs, id)
			s.queueMu.Unlock()
		}
	}
}
