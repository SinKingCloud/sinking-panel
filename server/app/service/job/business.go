package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"server/app/enum/system_task_status"
	cursorLog "server/app/util/log"
	"strconv"
	"strings"
	"time"
)

// Info 获取任务
func (s *service) Info(id string) *Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record := s.records[id]
	if record == nil {
		return nil
	}
	snapshot := *record
	return &snapshot
}

// List 获取任务列表
func (s *service) List(status *int) ([]*Job, error) {
	if status != nil {
		if _, ok := system_task_status.Map()[*status]; !ok {
			return nil, errors.New("任务状态参数不合法")
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]*Job, 0, len(s.records))
	for _, record := range s.records {
		if status != nil && record.Status != *status {
			continue
		}
		snapshot := *record
		records = append(records, &snapshot)
	}
	return records, nil
}

// Create 创建任务
func (s *service) Create(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string))) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.closed {
		return errors.New("系统任务服务已停止")
	}
	now := time.Now().Unix()
	ctx, cancel := context.WithCancel(s.ctx)
	record := &Job{
		ID:         id,
		Name:       name,
		Status:     system_task_status.Pending,
		Progress:   0,
		Data:       data,
		CreateTime: now,
		UpdateTime: now,
		cancel:     cancel,
	}
	s.mu.Lock()
	if s.records[id] != nil {
		s.mu.Unlock()
		cancel()
		return errors.New("任务ID已存在")
	}
	s.records[id] = record
	s.mu.Unlock()
	s.appendLog(id, system_task_status.Pending, 0, "任务已创建，等待执行")

	s.queueMu.Lock()
	s.jobs[id] = &queuedJob{
		id:   id,
		ctx:  ctx,
		data: data,
		run:  run,
	}
	s.queue = append(s.queue, id)
	s.queueCond.Signal()
	s.queueMu.Unlock()
	return nil
}

// Update 更新任务
func (s *service) Update(id string, status int, progress float64, message string, result ...interface{}) {
	if _, ok := system_task_status.Map()[status]; !ok {
		return
	}
	s.mu.Lock()

	record, ok := s.records[id]
	if !ok {
		s.mu.Unlock()
		return
	}
	terminal := record.Status == system_task_status.Completed || record.Status == system_task_status.Failed || record.Status == system_task_status.Canceled
	if terminal && record.Status != status {
		// 取消和执行结束可能同时发生；保持取消状态，也保留 worker 的最终统计。
		if record.Status == system_task_status.Canceled && len(result) > 0 {
			record.Result = result[0]
		}
		s.mu.Unlock()
		return
	}
	statusChanged := record.Status != status
	messageChanged := record.Message != message
	progressChanged := record.Progress != progress
	now := time.Now()
	shouldLog := statusChanged || messageChanged
	finalStatus := status == system_task_status.Completed || status == system_task_status.Failed || status == system_task_status.Canceled
	if progressChanged && (record.lastLogAt.IsZero() || now.Sub(record.lastLogAt) >= time.Second || finalStatus) {
		shouldLog = true
	}

	record.Status = status
	record.Progress = progress
	record.Message = message
	record.UpdateTime = now.Unix()

	if status == system_task_status.Running && record.StartTime == 0 {
		record.StartTime = record.UpdateTime
	}

	if finalStatus {
		record.EndTime = record.UpdateTime
		if len(result) > 0 {
			record.Result = result[0]
		}
	}
	if shouldLog {
		record.lastLogAt = now
	}
	s.mu.Unlock()
	if shouldLog {
		logMessage := message
		if progressChanged && strings.TrimSpace(logMessage) == "" {
			logMessage = "进度更新"
		}
		s.appendLog(id, status, progress, logMessage)
	}
}

// Cancel 取消任务
func (s *service) Cancel(id string) bool {
	s.mu.Lock()
	record, ok := s.records[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	// 等待中和运行中的任务都可以取消。
	if record.Status != system_task_status.Pending && record.Status != system_task_status.Running {
		s.mu.Unlock()
		return false
	}
	cancel := record.cancel
	record.cancel = nil
	// 更新任务状态
	record.Status = system_task_status.Canceled
	record.Message = "任务已取消"
	record.UpdateTime = time.Now().Unix()
	record.EndTime = record.UpdateTime
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.queueMu.Lock()
	delete(s.jobs, id)
	s.queueMu.Unlock()
	s.appendLog(id, system_task_status.Canceled, 0, "任务已取消")
	return true
}

// Delete 取消仍在运行的工作，再删除记录及日志。
func (s *service) Delete(id string) bool {
	if s.Cancel(id) {
		s.mu.Lock()
		delete(s.records, id)
		s.mu.Unlock()
		s.removeLog(id)
		return true
	}

	s.mu.Lock()
	record, ok := s.records[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	cancel := record.cancel
	delete(s.records, id)
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	s.queueMu.Lock()
	delete(s.jobs, id)
	s.queueMu.Unlock()
	s.removeLog(id)
	return true
}

// worker 从共享队列取任务。worker 数量固定，任务数量不会直接创建同等数量的 goroutine。
func (s *service) worker() {
	defer s.wait.Done()
	for {
		job, ok := s.next()
		if !ok {
			return
		}
		s.run(job)
	}
}

func (s *service) next() (*queuedJob, bool) {
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

func (s *service) run(job *queuedJob) {
	if job == nil {
		return
	}

	s.mu.Lock()
	record, ok := s.records[job.id]
	if !ok || record.Status != system_task_status.Pending || job.ctx.Err() != nil {
		s.mu.Unlock()
		return
	}
	record.Status = system_task_status.Running
	record.StartTime = time.Now().Unix()
	record.UpdateTime = record.StartTime
	s.mu.Unlock()
	s.appendLog(job.id, system_task_status.Running, 0, "任务开始执行")

	update := func(status int, progress float64, message string) {
		s.Update(job.id, status, progress, message)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			s.Update(job.id, system_task_status.Failed, 0, "任务执行异常: "+fmt.Sprint(recovered))
		}
		s.finish(job)
	}()

	if job.run == nil {
		s.Update(job.id, system_task_status.Failed, 0, "任务执行器为空")
		return
	}
	job.run(job.ctx, job.data, update)
}

func (s *service) finish(job *queuedJob) {
	s.mu.Lock()
	record, ok := s.records[job.id]
	status := system_task_status.Failed
	var cancel context.CancelFunc
	if ok {
		status = record.Status
		cancel, record.cancel = record.cancel, nil
	}
	s.mu.Unlock()
	if cancel != nil {
		defer cancel()
	}
	terminal := status == system_task_status.Completed || status == system_task_status.Failed || status == system_task_status.Canceled
	if !ok || terminal {
		return
	}
	if job.ctx.Err() != nil {
		s.Update(job.id, system_task_status.Canceled, 0, "任务已取消")
		return
	}
	s.Update(job.id, system_task_status.Failed, 0, "任务未正常结束")
}

func (s *service) logPath(id string) string {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\\`) {
		return ""
	}
	return filepath.Join(s.logDirectory, id+".log")
}

func (s *service) appendLog(id string, status int, progress float64, message string) {
	fileName := s.logPath(id)
	if fileName == "" {
		return
	}
	message = strings.TrimSpace(strings.TrimRight(message, "\r\n"))
	if message == "" {
		return
	}
	statusName := system_task_status.Map()[status]
	line := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	if statusName != "" {
		line += " [" + statusName + "]"
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	progressText := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(progress, 'f', 2, 64), "0"), ".")
	if progressText == "" {
		progressText = "0"
	}
	line += " [" + progressText + "%] " + message + "\n"

	s.logMu.Lock()
	defer s.logMu.Unlock()
	if err := os.MkdirAll(s.logDirectory, 0755); err != nil {
		return
	}
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

func (s *service) removeLog(id string) {
	fileName := s.logPath(id)
	if fileName == "" {
		return
	}
	s.logMu.Lock()
	defer s.logMu.Unlock()
	_ = os.Remove(fileName)
}

// Log 读取日志，支持读取最新行及按游标向前、向后翻页。
func (s *service) Log(id string, after, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	return cursorLog.Read(s.logPath(id), after, before, pageSize)
}
