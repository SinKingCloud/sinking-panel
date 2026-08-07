package task

import (
	"bytes"
	"errors"
	"io"
	"os"
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
	if err := s.repositoryTask.DeleteByIds(ids); err != nil {
		return err
	}
	for _, task := range tasks {
		if task.EntryID > 0 {
			s.instance.Remove(cron.EntryID(task.EntryID))
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

// getTaskLogFile 获取任务日志文件
func (s *service) getTaskLogFile(id int64) string {
	name := "task-" + strconv.FormatInt(id, 10) + ".log"
	f := file.NewDisk(s.getTaskLogPath())
	if !f.Exists(name) {
		_ = f.AutoCreate(name)
	}
	return strings.ReplaceAll(s.getTaskLogPath()+name, "//", "")
}

// ReadLog 读取日志，第一页返回最新输出，后续页按时间倒序读取历史日志。
func (s *service) ReadLog(id int64, page int, pageSize int) []string {
	s.logLock.RLock()
	defer s.logLock.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10000
	}
	fileName := s.getTaskLogFile(id)
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer func() {
		_ = f.Close()
	}()

	stat, err := f.Stat()
	if err != nil || stat.Size() == 0 {
		return nil
	}

	// 从文件末尾按块读取，避免日志越大轮询成本越高。
	const chunkSize int64 = 64 * 1024
	targetLines := page * pageSize
	position := stat.Size()
	newLineCount := 0
	chunks := make([][]byte, 0, pageSize+1)
	totalSize := 0
	for position > 0 && newLineCount <= targetLines {
		readSize := chunkSize
		if position < readSize {
			readSize = position
		}
		position -= readSize
		chunk := make([]byte, int(readSize))
		_, readErr := f.ReadAt(chunk, position)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil
		}
		chunks = append(chunks, chunk)
		totalSize += len(chunk)
		newLineCount += bytes.Count(chunk, []byte{'\n'})
	}

	data := make([]byte, 0, totalSize)
	for i := len(chunks) - 1; i >= 0; i-- {
		data = append(data, chunks[i]...)
	}
	parts := bytes.Split(data, []byte{'\n'})
	if position > 0 && len(parts) > 0 {
		parts = parts[1:]
	}
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}
	offset := (page - 1) * pageSize
	if offset >= len(parts) {
		return []string{}
	}
	pageEnd := len(parts) - offset
	pageStart := pageEnd - pageSize
	if pageStart < 0 {
		pageStart = 0
	}
	parts = parts[pageStart:pageEnd]
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		lines = append(lines, strings.TrimSuffix(string(part), "\r"))
	}
	return lines
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

	fileName := s.getTaskLogFile(id)
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
