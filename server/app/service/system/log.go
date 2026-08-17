package system

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"server/app/enum/system_task_status"
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

// resetTaskLogs removes logs left by a previous process. System tasks are
// intentionally in-memory, so a stale log must not be presented as a live task.
func (s *service) resetTaskLogs() {
	_ = os.RemoveAll(s.systemTaskLogDirectory)
	_ = os.MkdirAll(s.systemTaskLogDirectory, 0755)
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
func (s *service) TaskLog(id string, after int64, before int64, pageSize int) map[string]interface{} {
	s.logMu.RLock()
	defer s.logMu.RUnlock()

	if pageSize < 1 {
		pageSize = 300
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	fileName := s.taskLogPath(id)
	result := map[string]interface{}{
		"file_name":    filepath.Base(fileName),
		"lines":        []string{},
		"cursor":       after,
		"start_cursor": after,
		"has_previous": after > 0,
		"end":          false,
	}
	if fileName == "" {
		result["file_name"] = ""
		return result
	}
	f, err := os.Open(fileName)
	if err != nil {
		return result
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil || stat.Size() == 0 {
		result["cursor"] = int64(0)
		result["start_cursor"] = int64(0)
		result["has_previous"] = false
		return result
	}

	readBefore := func(end int64) ([]string, []int64) {
		if end < 0 {
			end = 0
		}
		if end > stat.Size() {
			end = stat.Size()
		}
		const chunkSize int64 = 64 * 1024
		position := end
		lineCount := 0
		chunks := make([][]byte, 0, 2)
		totalSize := 0
		for position > 0 && lineCount <= pageSize {
			readSize := chunkSize
			if position < readSize {
				readSize = position
			}
			position -= readSize
			chunk := make([]byte, int(readSize))
			_, readErr := f.ReadAt(chunk, position)
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return nil, nil
			}
			chunks = append(chunks, chunk)
			totalSize += len(chunk)
			lineCount += bytes.Count(chunk, []byte{'\n'})
		}
		data := make([]byte, 0, totalSize)
		for i := len(chunks) - 1; i >= 0; i-- {
			data = append(data, chunks[i]...)
		}
		start := 0
		if position > 0 {
			lineEnd := bytes.IndexByte(data, '\n')
			if lineEnd < 0 {
				return nil, nil
			}
			start = lineEnd + 1
		}
		lines := make([]string, 0, pageSize)
		starts := make([]int64, 0, pageSize)
		for start < len(data) {
			relEnd := bytes.IndexByte(data[start:], '\n')
			if relEnd < 0 {
				if end == stat.Size() {
					lines = append(lines, strings.TrimSuffix(string(data[start:]), "\r"))
					starts = append(starts, position+int64(start))
				}
				break
			}
			lineEnd := start + relEnd
			lines = append(lines, strings.TrimSuffix(string(data[start:lineEnd]), "\r"))
			starts = append(starts, position+int64(start))
			start = lineEnd + 1
		}
		if len(lines) > pageSize {
			first := len(lines) - pageSize
			lines = lines[first:]
			starts = starts[first:]
		}
		return lines, starts
	}

	readAfter := func(start int64) ([]string, int64) {
		if start < 0 {
			start = 0
		}
		if start >= stat.Size() {
			return []string{}, start
		}
		if _, err = f.Seek(start, io.SeekStart); err != nil {
			return nil, start
		}
		reader := bufio.NewReaderSize(f, 64*1024)
		lines := make([]string, 0, pageSize)
		next := start
		for len(lines) < pageSize {
			line, readErr := reader.ReadString('\n')
			if len(line) == 0 && readErr != nil {
				break
			}
			if len(line) == 0 || line[len(line)-1] != '\n' {
				break
			}
			next += int64(len(line))
			lines = append(lines, strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"))
			if readErr != nil {
				break
			}
		}
		return lines, next
	}

	if after > 0 {
		if after > stat.Size() {
			lines, starts := readBefore(stat.Size())
			result["lines"] = lines
			result["cursor"] = stat.Size()
			result["end"] = true
			if len(starts) > 0 {
				result["start_cursor"] = starts[0]
				result["has_previous"] = starts[0] > 0
			}
			return result
		}
		lines, next := readAfter(after)
		result["lines"] = lines
		result["cursor"] = next
		result["start_cursor"] = after
		result["has_previous"] = after > 0
		result["end"] = next >= stat.Size()
		return result
	}

	end := stat.Size()
	if before > 0 {
		end = before
	}
	lines, starts := readBefore(end)
	result["lines"] = lines
	result["cursor"] = stat.Size()
	if len(starts) > 0 {
		result["start_cursor"] = starts[0]
		result["has_previous"] = starts[0] > 0
		result["end"] = starts[0] <= 0
	} else {
		result["start_cursor"] = int64(0)
		result["has_previous"] = false
		result["end"] = true
	}
	return result
}
