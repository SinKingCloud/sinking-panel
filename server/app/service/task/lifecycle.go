package task

import "github.com/robfig/cron/v3"

// removeEntry 在 taskLock 内移除调度，已执行的旧快照保留至执行结束。
func (s *service) removeEntry(id cron.EntryID) {
	if id <= 0 {
		return
	}
	j, _ := s.instance.Entry(id).Job.(*job)
	s.instance.Remove(id)
	if j == nil {
		return
	}
	if j.entryID > 0 && j.entryID != id {
		return
	}
	j.entryID = 0
	if j.Id >= 0 {
		j.retired = true
		if j.running == 0 {
			j.cancel()
			delete(s.jobs, j)
		}
	}
}

// cancelJobs 在 taskLock 内取消匹配的执行和调度，不修改任务配置与日志。
func (s *service) cancelJobs(match func(*job) bool) []*job {
	jobs := make([]*job, 0)
	for j := range s.jobs {
		if !match(j) {
			continue
		}
		j.retired = true
		j.cancel()
		if j.entryID > 0 {
			s.instance.Remove(j.entryID)
			j.entryID = 0
		}
		jobs = append(jobs, j)
	}
	return jobs
}

// waitJobs 不持有任务锁等待回调，回调仍可写日志或读取任务信息。
func (s *service) waitJobs(jobs []*job) {
	for _, j := range jobs {
		j.wait.Wait()
		s.taskLock.Lock()
		delete(s.jobs, j)
		if s.internalTasks[j.Id] == j {
			delete(s.internalTasks, j.Id)
		}
		s.taskLock.Unlock()
	}
}

// Cancel 按 ID 注销任务的所有调度和执行，保留数据库配置和日志。
func (s *service) Cancel(id int64) error {
	s.taskLock.Lock()
	jobs := s.cancelJobs(func(j *job) bool { return j.Id == id })
	s.taskLock.Unlock()
	s.waitJobs(jobs)
	return nil
}
