package system

import (
	"server/app/enum/system_task_status"
	"strings"
	"time"
)

// TaskUpdate 更新任务
func (s *service) TaskUpdate(id string, status int, progress float64, message string) {
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
