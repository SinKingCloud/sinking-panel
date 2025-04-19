package task

import (
	"bufio"
	"errors"
	"github.com/robfig/cron/v3"
	"os"
	"server/app/constant"
	"server/app/model"
	"server/app/util/file"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	instance = cron.New(
		cron.WithSeconds(),
		cron.WithChain(
			cron.Recover(cron.DiscardLogger),
		),
	)
	instanceOnce = &sync.Once{}
)

// entryId 获取entryId
func (s *Service) entryId(id int) (cron.EntryID, error) {
	task, err := s.findById(id)
	if err != nil {
		return 0, err
	}
	if task.EntryID <= 0 {
		return 0, errors.New("任务未实例化")
	}
	entry := instance.Entry(cron.EntryID(task.EntryID))
	if entry.ID <= 0 {
		return 0, errors.New("未查询到对应实例")
	}
	return entry.ID, nil
}

// Stop 暂停任务
func (s *Service) Stop(id int) error {
	entryID, err := s.entryId(id)
	if err == nil && int(entryID) > 0 {
		instance.Remove(entryID)
	}
	_ = s.updateEntryIDById(id, 0)
	return s.updateStatusById(id, Stop)
}

// Restore 恢复任务
func (s *Service) Restore(id int) error {
	task, err := s.findById(id)
	if err != nil {
		return err
	}
	if task.EntryID > 0 {
		instance.Remove(cron.EntryID(task.EntryID))
	}
	j := newJob(task)
	if j == nil {
		return errors.New("任务实例化失败")
	}
	entryID, err := instance.AddJob(task.Spec, j)
	if err == nil && int(entryID) > 0 {
		_ = s.updateEntryIDById(id, int(entryID))
	}
	return s.updateStatusById(id, Running)
}

// Run 执行任务
func (s *Service) Run(id int) error {
	entryID, err := s.entryId(id)
	if err != nil {
		return err
	}
	entry := instance.Entry(entryID)
	if entry.ID > 0 {
		go entry.Job.Run()
	}
	return nil
}

// Remove 移除任务
func (s *Service) Remove(id int) error {
	entryID, err := s.entryId(id)
	if err == nil && int(entryID) > 0 {
		instance.Remove(entryID)
	}
	return s.deleteById(id)
}

// Refresh 刷新任务
func (s *Service) Refresh(id int) error {
	entryID, err := s.entryId(id)
	if err == nil && int(entryID) > 0 {
		instance.Remove(entryID)
	}
	task, err := s.findById(id)
	if err != nil {
		return err
	}
	if task.Status == int(Running) {
		j := newJob(task)
		entryId, err := instance.AddJob(task.Spec, j)
		if err != nil {
			return err
		}
		return s.updateEntryIDById(id, int(entryId))
	}
	return nil
}

// Add 添加任务
func (s *Service) Add(data *model.Task) error {
	j := newJob(data)
	if j == nil {
		return errors.New("任务实例化失败")
	}
	id, err := instance.AddJob(data.Spec, j)
	if err != nil {
		return err
	}
	data.EntryID = int(id)
	err = s.create(data)
	if err != nil {
		instance.Remove(id)
	}
	return err
}

// Start 启动任务
func (s *Service) Start() {
	go instanceOnce.Do(func() {
		tasks, err := s.selectAll()
		if err == nil && tasks != nil {
			for _, task := range tasks {
				if task.Status != int(Running) {
					continue
				}
				job2 := newJob(task)
				if job2 != nil {
					entryID, err2 := instance.AddJob(task.Spec, job2)
					if err2 == nil && entryID > 0 {
						_ = s.updateEntryIDById(task.Id, int(entryID))
					}
				}
			}
		}
		instance.Start()
	})
}

// getTaskLogFile 获取任务日志文件
func (s *Service) getTaskLogPath() string {
	path := constant.TempPath + "/cron"
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return strings.ReplaceAll(path, "//", "")
}

// getTaskLogFile 获取任务日志文件
func (s *Service) getTaskLogFile(id int) string {
	name := "task-" + strconv.Itoa(id) + ".log"
	f := file.NewDisk(s.getTaskLogPath())
	if !f.Exists(name) {
		_ = f.AutoCreate(name)
	}
	return strings.ReplaceAll(s.getTaskLogPath()+name, "//", "")
}

// GetLog 获取log信息
func (s *Service) getLines(filename string, page int, pageSize int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	scanner := bufio.NewScanner(f)
	skipLines := (page - 1) * pageSize
	for i := 0; i < skipLines; i++ {
		if !scanner.Scan() {
			return nil, nil // 没有更多行可读取
		}
	}
	var lines []string
	for i := 0; i < pageSize; i++ {
		if !scanner.Scan() {
			break
		}
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// ReadLog 读取日志
func (s *Service) ReadLog(id int, page int, pageSize int) []string {
	l, _ := s.getLines(s.getTaskLogFile(id), page, pageSize)
	return l
}

// WriteLog 写入日志
func (s *Service) WriteLog(id int, content string) error {
	fileName := s.getTaskLogFile(id)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("[" + time.Now().Format("2006-01-02 15:04:05") + "] " + content + "\n")
	return err
}

// ValidateCron 判断cron表达式
func (s *Service) ValidateCron(expr string) bool {
	parser := cron.NewParser(
		cron.SecondOptional |
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
