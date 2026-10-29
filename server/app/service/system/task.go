package system

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

// GetTask 获取任务
func (s *service) GetTask(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task := s.tasks[id]
	if task == nil {
		return nil
	}
	snapshot := *task
	return &snapshot
}

// TaskList 获取任务列表
func (s *service) TaskList(status *int) ([]*Task, error) {
	if status != nil {
		if _, ok := system_task_status.Map()[*status]; !ok {
			return nil, errors.New("任务状态参数不合法")
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if status != nil && task.Status != *status {
			continue
		}
		snapshot := *task
		tasks = append(tasks, &snapshot)
	}
	return tasks, nil
}

// TaskCreate 创建任务
func (s *service) TaskCreate(id, name string, data interface{}, run func(context.Context, interface{}, func(int, float64, string))) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.closed {
		return
	}
	now := time.Now().Unix()
	ctx, cancel := context.WithCancel(s.ctx)
	task := &Task{
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
	s.tasks[id] = task
	s.mu.Unlock()
	s.appendTaskLog(id, system_task_status.Pending, 0, "任务已创建，等待执行")

	s.queueMu.Lock()
	s.jobs[id] = &job{
		id:   id,
		ctx:  ctx,
		data: data,
		run:  run,
	}
	s.queue = append(s.queue, id)
	s.queueCond.Signal()
	s.queueMu.Unlock()
}

// TaskUpdate 更新任务
func (s *service) TaskUpdate(id string, status int, progress float64, message string, result ...interface{}) {
	if _, ok := system_task_status.Map()[status]; !ok {
		return
	}
	s.mu.Lock()

	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return
	}
	terminal := task.Status == system_task_status.Completed || task.Status == system_task_status.Failed || task.Status == system_task_status.Canceled
	if terminal && task.Status != status {
		// 取消和执行结束可能同时发生；保持取消状态，也保留 worker 的最终统计。
		if task.Status == system_task_status.Canceled && len(result) > 0 {
			task.Result = result[0]
		}
		s.mu.Unlock()
		return
	}
	statusChanged := task.Status != status
	messageChanged := task.Message != message
	progressChanged := task.Progress != progress
	now := time.Now()
	shouldLog := statusChanged || messageChanged
	finalStatus := status == system_task_status.Completed || status == system_task_status.Failed || status == system_task_status.Canceled
	if progressChanged && (task.lastLogAt.IsZero() || now.Sub(task.lastLogAt) >= time.Second || finalStatus) {
		shouldLog = true
	}

	task.Status = status
	task.Progress = progress
	task.Message = message
	task.UpdateTime = now.Unix()

	if status == system_task_status.Running && task.StartTime == 0 {
		task.StartTime = task.UpdateTime
	}

	if finalStatus {
		task.EndTime = task.UpdateTime
		if len(result) > 0 {
			task.Result = result[0]
		}
	}
	if shouldLog {
		task.lastLogAt = now
	}
	s.mu.Unlock()
	if shouldLog {
		logMessage := message
		if progressChanged && strings.TrimSpace(logMessage) == "" {
			logMessage = "进度更新"
		}
		s.appendTaskLog(id, status, progress, logMessage)
	}
}

// TaskCancel 取消任务
func (s *service) TaskCancel(id string) bool {
	s.mu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return false
	}
	// 等待中和运行中的任务都可以取消。
	if task.Status != system_task_status.Pending && task.Status != system_task_status.Running {
		s.mu.Unlock()
		return false
	}
	cancel := task.cancel
	// 更新任务状态
	task.Status = system_task_status.Canceled
	task.Message = "任务已取消"
	task.UpdateTime = time.Now().Unix()
	task.EndTime = task.UpdateTime
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.queueMu.Lock()
	delete(s.jobs, id)
	s.queueMu.Unlock()
	s.appendTaskLog(id, system_task_status.Canceled, 0, "任务已取消")
	return true
}

// TaskDelete cancels an active task or removes a completed task and its log.
func (s *service) TaskDelete(id string) bool {
	// Always cancel first. This invalidates the worker context and removes a
	// queued job before the task record and its log are deleted.
	if s.TaskCancel(id) {
		s.mu.Lock()
		delete(s.tasks, id)
		s.mu.Unlock()
		s.removeTaskLog(id)
		return true
	}

	s.mu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return false
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
	s.removeTaskLog(id)
	return true
}

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

func (s *service) taskLogPath(id string) string {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\\`) {
		return ""
	}
	return filepath.Join(s.systemTaskLogDirectory, id+".log")
}

func (s *service) appendTaskLog(id string, status int, progress float64, message string) {
	fileName := s.taskLogPath(id)
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
	if err := os.MkdirAll(s.systemTaskLogDirectory, 0755); err != nil {
		return
	}
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

func (s *service) removeTaskLog(id string) {
	fileName := s.taskLogPath(id)
	if fileName == "" {
		return
	}
	s.logMu.Lock()
	defer s.logMu.Unlock()
	_ = os.Remove(fileName)
}

// TaskLog reads the newest lines, incremental lines, or older lines according
// to the same byte-cursor contract used by scheduled-task logs.
func (s *service) TaskLog(id string, after, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	return cursorLog.Read(s.taskLogPath(id), after, before, pageSize)
}
