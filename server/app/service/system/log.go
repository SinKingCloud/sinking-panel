package system

import (
	"os"
	"path/filepath"
	"server/app/enum/system_task_status"
	cursorLog "server/app/util/log"
	"strconv"
	"strings"
	"time"
)

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
func (s *service) TaskLog(id string, after int64, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	return cursorLog.Read(s.taskLogPath(id), after, before, pageSize)
}
