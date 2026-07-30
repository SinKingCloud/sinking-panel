package task

import (
	"bufio"
	"errors"
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

// entryId 获取任务实例ID
func (s *service) entryId(id int64) (cron.EntryID, error) {
	task, err := s.findById(id)
	if err != nil {
		return 0, err
	}
	if task.EntryID <= 0 {
		return 0, errors.New("任务未实例化")
	}
	entry := s.instance.Entry(cron.EntryID(task.EntryID))
	if entry.ID <= 0 {
		return 0, errors.New("未查询到对应实例")
	}
	return entry.ID, nil
}

// Stop 暂停任务
func (s *service) Stop(id int64) error {
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
	entryID, err := s.entryId(id)
	if err != nil {
		return err
	}
	entry := s.instance.Entry(entryID)
	if entry.ID > 0 {
		go entry.Job.Run()
	}
	return nil
}

// Remove 移除任务
func (s *service) Remove(ids []int64) error {
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

// Refresh 刷新任务
func (s *service) Refresh(id int64) error {
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

// Add 添加任务
func (s *service) Add(data *model.Task) error {
	if data == nil {
		return errors.New("任务数据不能为空")
	}
	data.Id = str.GetSnowWorkIns().GetId()
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

// GetLog 获取log信息
func (s *service) getLines(filename string, page int, pageSize int) ([]string, error) {
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
func (s *service) ReadLog(id int64, page int, pageSize int) []string {
	l, _ := s.getLines(s.getTaskLogFile(id), page, pageSize)
	return l
}

// WriteLog 写入日志
func (s *service) WriteLog(id int64, content string) error {
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
func (s *service) ValidateCron(expr string) bool {
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
