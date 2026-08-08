package task

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/enum/task_status"
	"server/app/model"
	"server/app/util/file"
	"server/app/util/str"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Stop 暂停任务
func (s *service) Stop(id int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	task, err := s.findById(id)
	if err != nil {
		return err
	}
	if err = s.updateStateById(id, 0, task_status.Stop); err != nil {
		return err
	}
	if task.EntryID > 0 {
		s.instance.Remove(cron.EntryID(task.EntryID))
	}
	return nil
}

// Restore 恢复任务
func (s *service) Restore(id int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	task, err := s.findById(id)
	if err != nil {
		return err
	}
	j := newJob(task, s)
	if j == nil {
		return errors.New("任务实例化失败")
	}
	entryID, err := s.instance.AddJob(task.Spec, j)
	if err != nil {
		return err
	}
	if entryID <= 0 {
		return errors.New("任务实例化失败")
	}
	if err = s.updateStateById(id, int(entryID), task_status.Running); err != nil {
		s.instance.Remove(entryID)
		return err
	}
	if task.EntryID > 0 && cron.EntryID(task.EntryID) != entryID {
		s.instance.Remove(cron.EntryID(task.EntryID))
	}
	return nil
}

// Run 执行任务
func (s *service) Run(id int64) error {
	s.taskLock.Lock()
	task, err := s.findById(id)
	if err != nil {
		s.taskLock.Unlock()
		return err
	}
	j := newJob(task, s)
	if j == nil {
		s.taskLock.Unlock()
		return errors.New("任务实例化失败")
	}
	s.taskLock.Unlock()
	go j.Run()
	return nil
}

// Remove 移除任务
func (s *service) Remove(ids []int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	if len(ids) == 0 {
		return errors.New("删除任务不能为空")
	}
	tasks := make([]*model.Task, 0, len(ids))
	for _, id := range ids {
		task, err := s.findById(id)
		if err != nil {
			return err
		}
		tasks = append(tasks, task)
	}
	s.logLock.Lock()
	defer s.logLock.Unlock()
	if err := s.repositoryTask.DeleteByIds(ids); err != nil {
		return err
	}
	for _, task := range tasks {
		if task.EntryID > 0 {
			s.instance.Remove(cron.EntryID(task.EntryID))
		}
		if err := os.Remove(s.getTaskLogFilePath(task.Id)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// refresh 刷新任务
func (s *service) refresh(id int64) error {
	task, err := s.findById(id)
	if err != nil {
		return err
	}
	oldEntryID := cron.EntryID(task.EntryID)
	if task.Status != task_status.Running {
		if oldEntryID > 0 {
			if err = s.updateEntryIDById(id, 0); err != nil {
				return err
			}
			s.instance.Remove(oldEntryID)
		}
		return nil
	}
	j := newJob(task, s)
	if j == nil {
		return errors.New("任务实例化失败")
	}
	entryID, err := s.instance.AddJob(task.Spec, j)
	if err != nil {
		return err
	}
	if entryID <= 0 {
		return errors.New("任务实例化失败")
	}
	if err = s.updateEntryIDById(id, int(entryID)); err != nil {
		s.instance.Remove(entryID)
		return err
	}
	if oldEntryID > 0 && oldEntryID != entryID {
		s.instance.Remove(oldEntryID)
	}
	return nil
}

// Refresh 刷新任务
func (s *service) Refresh(id int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	return s.refresh(id)
}

// Add 添加任务
func (s *service) Add(data *model.Task) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	if data == nil {
		return errors.New("任务数据不能为空")
	}
	if _, ok := task_status.Map()[data.Status]; !ok {
		return errors.New("任务状态不合法")
	}
	data.Id = str.GetSnowWorkIns().GetId()
	if data.Status == task_status.Stop {
		data.EntryID = 0
		return s.create(data)
	}
	j := newJob(data, s)
	if j == nil {
		return errors.New("任务实例化失败")
	}
	id, err := s.instance.AddJob(data.Spec, j)
	if err != nil {
		return err
	}
	data.EntryID = int(id)
	err = s.create(data)
	if err != nil {
		s.instance.Remove(id)
	}
	return err
}

// Start 启动任务
func (s *service) Start() {
	go s.startOnce.Do(func() {
		s.taskLock.Lock()
		defer s.taskLock.Unlock()

		tasks, err := s.selectAll()
		if err == nil && tasks != nil {
			for _, task := range tasks {
				if task.Status != task_status.Running {
					continue
				}
				job2 := newJob(task, s)
				if job2 != nil {
					entryID, err2 := s.instance.AddJob(task.Spec, job2)
					if err2 == nil && entryID > 0 {
						if err2 = s.updateEntryIDById(task.Id, int(entryID)); err2 != nil {
							s.instance.Remove(entryID)
						}
					}
				}
			}
		}
		s.instance.Start()
	})
}

// getTaskLogFile 获取任务日志文件
func (s *service) getTaskLogPath() string {
	path := constant.TempPath + "/cron"
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return strings.ReplaceAll(path, "//", "")
}

func (s *service) getTaskLogFilePath(id int64) string {
	name := "task-" + strconv.FormatInt(id, 10) + ".log"
	return strings.ReplaceAll(s.getTaskLogPath()+name, "//", "")
}

// getTaskLogFile 获取任务日志文件
func (s *service) getTaskLogFile(id int64) string {
	name := "task-" + strconv.FormatInt(id, 10) + ".log"
	f := file.NewDisk(s.getTaskLogPath())
	if !f.Exists(name) {
		_ = f.AutoCreate(name)
	}
	return s.getTaskLogFilePath(id)
}

// ReadLog 读取日志。首次读取最新内容，cursor 读取新增内容，before 读取更早内容。
func (s *service) ReadLog(id int64, cursor int64, before int64, pageSize int) map[string]interface{} {
	s.logLock.RLock()
	defer s.logLock.RUnlock()

	if pageSize < 1 {
		pageSize = 300
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	fileName := s.getTaskLogFile(id)
	f, err := os.Open(fileName)
	if err != nil {
		return map[string]interface{}{
			"file_name":    filepath.Base(fileName),
			"lines":        []string{},
			"cursor":       cursor,
			"start_cursor": cursor,
			"has_previous": false,
			"end":          false,
		}
	}
	defer func() {
		_ = f.Close()
	}()

	stat, err := f.Stat()
	if err != nil {
		return map[string]interface{}{
			"file_name":    filepath.Base(fileName),
			"lines":        []string{},
			"cursor":       cursor,
			"start_cursor": cursor,
			"has_previous": false,
			"end":          false,
		}
	}
	result := map[string]interface{}{
		"file_name":    filepath.Base(fileName),
		"lines":        []string{},
		"cursor":       cursor,
		"start_cursor": cursor,
		"has_previous": cursor > 0,
		"end":          false,
	}
	if stat.Size() == 0 {
		result["cursor"] = int64(0)
		result["start_cursor"] = int64(0)
		result["has_previous"] = false
		return result
	}

	// 从文件尾部读取一段完整日志，并记录每一行的文件起始位置。
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

	if cursor > 0 {
		if cursor > stat.Size() {
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
		lines, next := readAfter(cursor)
		result["lines"] = lines
		result["cursor"] = next
		result["start_cursor"] = cursor
		result["has_previous"] = cursor > 0
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
	} else {
		result["start_cursor"] = int64(0)
		result["has_previous"] = false
	}
	return result
}

// ClearLog 清理任务日志。
func (s *service) ClearLog(id int64) error {
	s.logLock.Lock()
	defer s.logLock.Unlock()

	fileName := s.getTaskLogFile(id)
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	return f.Close()
}

// WriteLog 写入日志
func (s *service) WriteLog(id int64, content string) error {
	s.logLock.Lock()
	defer s.logLock.Unlock()

	// 统一清理调用方携带的行结束符，确保每条记录只追加一个换行。
	content = strings.TrimRight(content, "\r\n")
	if content == "" {
		return nil
	}

	fileName := s.getTaskLogFilePath(id)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if _, findErr := s.repositoryTask.FindById(id); findErr != nil {
			return nil
		}
	}
	fileName = s.getTaskLogFile(id)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.WriteString("[" + time.Now().Format("2006-01-02 15:04:05") + "] " + content + "\n")
	return err
}

// ValidateCron 判断cron表达式
func (s *service) ValidateCron(expr string) bool {
	parser := cron.NewParser(
		cron.Second |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow |
			cron.Descriptor,
	)
	_, err := parser.Parse(expr)
	return err == nil
}
