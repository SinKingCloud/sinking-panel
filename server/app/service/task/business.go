package task

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/enum/task_exec_type"
	"server/app/enum/task_status"
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/service/task/request"
	"server/app/util/file"
	cursorLog "server/app/util/log"
	"server/app/util/str"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// validateTypeId 校验任务分类
func (s *service) validateTypeId(typeId int64) error {
	if typeId < 0 {
		return errors.New("任务分类不合法")
	}
	if typeId == 0 {
		return nil
	}
	if s.typeService == nil {
		return errors.New("分类服务未初始化")
	}
	data, err := s.typeService.FindById(typeId)
	if err != nil || data == nil || data.Module != type_module.Task {
		return errors.New("任务分类不合法")
	}
	return nil
}

// checkContent 由对应任务类型校验并规范化任务内容。
func (s *service) checkContent(content string, execType int) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("任务内容不能为空")
	}
	switch execType {
	case task_exec_type.Script:
		return s.script.Validate(content)
	case task_exec_type.Request:
		return s.request.Validate(content)
	default:
		return "", errors.New("任务类型不合法")
	}
}

// formatContent 格式化任务内容。
func (s *service) formatContent(content string, execType int) interface{} {
	switch execType {
	case task_exec_type.Script:
		return content
	case task_exec_type.Request:
		value, err := s.request.Validate(content)
		if err == nil {
			data := request.Config{}
			if json.Unmarshal([]byte(value), &data) == nil {
				return data
			}
		}
		return request.Config{}
	default:
		return nil
	}
}

// Stop 暂停任务
func (s *service) Stop(id int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	task, err := s.findById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	entryID := task.EntryID
	if id < 0 {
		task.EntryID = 0
		task.Status = task_status.Stop
		task.UpdateTime = str.DateTime(time.Now())
	} else if err = s.repositoryTask.UpdateStateById(id, 0, task_status.Stop); err != nil {
		return err
	}
	if entryID > 0 {
		s.removeEntry(cron.EntryID(entryID))
	}
	return nil
}

// Restore 恢复任务
func (s *service) Restore(id int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	if s.closed {
		return errors.New("计划任务服务已关闭")
	}

	task, err := s.findById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	j := s.internalTasks[id]
	if j == nil {
		j = s.newJob(task)
	}
	if j == nil || j.retired {
		return errors.New("任务实例化失败")
	}
	defer func() {
		if id >= 0 && j.entryID == 0 && j.running == 0 {
			j.cancel()
			delete(s.jobs, j)
		}
	}()
	entryID, err := s.instance.AddJob(task.Spec, j)
	if err != nil {
		return err
	}
	if entryID <= 0 {
		return errors.New("任务实例化失败")
	}
	oldEntryID := cron.EntryID(task.EntryID)
	if id < 0 {
		task.EntryID = int(entryID)
		task.Status = task_status.Running
		task.UpdateTime = str.DateTime(time.Now())
	} else if err = s.repositoryTask.UpdateStateById(id, int(entryID), task_status.Running); err != nil {
		s.removeEntry(entryID)
		return err
	}
	j.entryID = entryID
	if oldEntryID > 0 && oldEntryID != entryID {
		s.removeEntry(oldEntryID)
	}
	return nil
}

// Run 执行任务
func (s *service) Run(id int64) error {
	s.taskLock.Lock()
	if s.closed {
		s.taskLock.Unlock()
		return errors.New("计划任务服务已关闭")
	}
	task, err := s.findById(id)
	if err != nil {
		s.taskLock.Unlock()
		return err
	}
	if task == nil {
		s.taskLock.Unlock()
		return errors.New("任务不存在")
	}
	j := s.internalTasks[id]
	if j == nil {
		j = s.newJob(task)
	}
	if j == nil || j.retired {
		s.taskLock.Unlock()
		return errors.New("任务实例化失败")
	}
	value, ok := j.begin()
	if !ok {
		s.taskLock.Unlock()
		return errors.New("任务已取消")
	}
	s.taskLock.Unlock()
	go func() {
		cron.Recover(cron.DiscardLogger)(cron.FuncJob(func() { j.execute(value) })).Run()
	}()
	return nil
}

// Remove 移除任务
func (s *service) Remove(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("删除任务不能为空")
	}
	s.taskLock.Lock()
	storedIds := make([]int64, 0, len(ids))
	selected := make(map[int64]bool, len(ids))
	for _, id := range ids {
		task, err := s.findById(id)
		if err != nil {
			s.taskLock.Unlock()
			return err
		}
		if task == nil {
			s.taskLock.Unlock()
			return errors.New("任务不存在")
		}
		selected[id] = true
		if id >= 0 {
			storedIds = append(storedIds, id)
		}
	}
	if len(storedIds) > 0 {
		if err := s.repositoryTask.DeleteByIds(storedIds); err != nil {
			s.taskLock.Unlock()
			return err
		}
	}
	jobs := s.cancelJobs(func(j *job) bool { return selected[j.Id] })
	s.taskLock.Unlock()
	s.waitJobs(jobs)
	s.logLock.Lock()
	defer s.logLock.Unlock()
	for _, id := range ids {
		if err := os.Remove(s.getTaskLogFilePath(id)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// refresh 刷新任务
func (s *service) refresh(id int64) error {
	if s.closed {
		return errors.New("计划任务服务已关闭")
	}
	task, err := s.findById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	oldEntryID := cron.EntryID(task.EntryID)
	if task.Status != task_status.Running {
		if oldEntryID > 0 {
			if id < 0 {
				task.EntryID = 0
			} else if err = s.repositoryTask.UpdateEntryIDById(id, 0); err != nil {
				return err
			}
			s.removeEntry(oldEntryID)
		}
		return nil
	}
	j := s.internalTasks[id]
	if j == nil {
		j = s.newJob(task)
	}
	if j == nil || j.retired {
		return errors.New("任务实例化失败")
	}
	defer func() {
		if id >= 0 && j.entryID == 0 && j.running == 0 {
			j.cancel()
			delete(s.jobs, j)
		}
	}()
	entryID, err := s.instance.AddJob(task.Spec, j)
	if err != nil {
		return err
	}
	if entryID <= 0 {
		return errors.New("任务实例化失败")
	}
	if id < 0 {
		task.EntryID = int(entryID)
	} else if err = s.repositoryTask.UpdateEntryIDById(id, int(entryID)); err != nil {
		s.removeEntry(entryID)
		return err
	}
	j.entryID = entryID
	if oldEntryID > 0 && oldEntryID != entryID {
		s.removeEntry(oldEntryID)
	}
	return nil
}

// Refresh 刷新任务
func (s *service) Refresh(id int64) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	return s.refresh(id)
}

// Add 添加任务。传入执行方法时使用固定负数 ID 注册内部任务，只保存在内存中。
func (s *service) Add(data *model.Task, run ...func(context.Context)) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	if s.closed {
		return errors.New("计划任务服务已关闭")
	}

	if data == nil {
		return errors.New("任务数据不能为空")
	}
	if len(run) > 1 || (len(run) == 1 && run[0] == nil) {
		return errors.New("任务执行方法不合法")
	}
	internal := len(run) == 1
	if internal {
		if data.Id >= 0 {
			return errors.New("内部任务必须使用固定负数ID")
		}
		if s.internalTasks[data.Id] != nil {
			return errors.New("内部任务ID已存在")
		}
	} else if data.Id < 0 {
		return errors.New("内部任务必须提供执行方法")
	}
	if err := s.validateTypeId(data.TypeId); err != nil {
		return err
	}
	if _, ok := task_exec_type.Map()[data.ExecType]; !ok {
		return errors.New("任务执行类型不合法")
	}
	if _, ok := task_status.Map()[data.Status]; !ok {
		return errors.New("任务状态不合法")
	}
	if strings.TrimSpace(data.Name) == "" {
		return errors.New("任务名称不能为空")
	}
	if !s.validateCron(data.Spec) {
		return errors.New("任务表达式不合法")
	}
	if internal {
		task := *data
		task.EntryID = 0
		task.CreateTime = str.DateTime(time.Now())
		task.UpdateTime = task.CreateTime
		id := task.Id
		j := &job{
			Task:    &task,
			service: s,
		}
		j.ctx, j.cancel = context.WithCancel(s.ctx)
		// 复用同一执行包装，改周期、暂停恢复和手动执行都不会重叠运行。
		j.run = cron.NewChain(
			cron.SkipIfStillRunning(cron.DiscardLogger),
			cron.Recover(cron.DiscardLogger),
		).Then(cron.FuncJob(func() {
			s.taskLock.Lock()
			active := !s.closed && !j.retired && s.internalTasks[id] == j
			s.taskLock.Unlock()
			if !active || j.ctx.Err() != nil {
				return
			}
			defer func() {
				s.taskLock.Lock()
				defer s.taskLock.Unlock()
				j.RunTime = str.DateTime(time.Now())
			}()
			run[0](j.ctx)
		}))
		if task.Status == task_status.Running {
			entryID, err := s.instance.AddJob(task.Spec, j)
			if err != nil {
				j.cancel()
				return err
			}
			task.EntryID = int(entryID)
			j.entryID = entryID
		}
		s.internalTasks[id] = j
		s.jobs[j] = struct{}{}
		*data = task
		return nil
	}
	content, err := s.checkContent(data.Script, data.ExecType)
	if err != nil {
		return err
	}
	data.Script = content
	data.EntryID = 0
	if data.Status == task_status.Stop {
		return s.repositoryTask.Create(data)
	}
	// 先保存并取得任务 ID，避免调度器在 ID 回填前执行任务。
	if err = s.repositoryTask.Create(data); err != nil {
		return err
	}
	j := s.newJob(data)
	if j == nil {
		return errors.Join(errors.New("任务实例化失败"), s.repositoryTask.DeleteByIds([]int64{data.Id}))
	}
	id, err := s.instance.AddJob(data.Spec, j)
	if err == nil {
		j.entryID = id
		err = s.repositoryTask.UpdateEntryIDById(data.Id, int(id))
		if err == nil {
			data.EntryID = int(id)
			return nil
		}
		s.removeEntry(id)
	}
	j.cancel()
	delete(s.jobs, j)
	return errors.Join(err, s.repositoryTask.DeleteByIds([]int64{data.Id}))
}

// Start 启动任务
func (s *service) Start() {
	s.startOnce.Do(func() {
		s.taskLock.Lock()
		defer s.taskLock.Unlock()
		if s.closed {
			return
		}

		tasks, err := s.repositoryTask.SelectAll()
		if err == nil && tasks != nil {
			for _, task := range tasks {
				if task.Status != task_status.Running {
					continue
				}
				// 业务初始化时可能已启用计划，启动时不重复注册同一调度任务。
				if current, ok := s.instance.Entry(cron.EntryID(task.EntryID)).Job.(*job); ok && current.Id == task.Id {
					continue
				}
				job2 := s.newJob(task)
				if job2 != nil {
					entryID, err2 := s.instance.AddJob(task.Spec, job2)
					if err2 == nil && entryID > 0 {
						if err2 = s.repositoryTask.UpdateEntryIDById(task.Id, int(entryID)); err2 != nil {
							s.removeEntry(entryID)
						} else {
							job2.entryID = entryID
						}
					}
					if job2.entryID == 0 {
						job2.cancel()
						delete(s.jobs, job2)
					}
				}
			}
		}
		s.instance.Start()
	})
}

// Close 停止计划任务并等待正在执行的任务结束。
func (s *service) Close() {
	s.closeOnce.Do(func() {
		s.taskLock.Lock()
		s.closed = true
		s.cancel()
		jobs := s.cancelJobs(func(*job) bool { return true })
		s.taskLock.Unlock()
		<-s.instance.Stop().Done()
		s.runWait.Wait()
		s.waitJobs(jobs)
	})
}

// getTaskLogPath 获取任务日志目录
func (s *service) getTaskLogPath() string {
	return filepath.Clean(constant.CronPath)
}

func (s *service) getTaskLogFilePath(id int64) string {
	name := "task-" + strconv.FormatInt(id, 10) + ".log"
	return filepath.Join(s.getTaskLogPath(), name)
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

// ReadLog 读取日志。首次读取最新内容，after 读取新增内容，before 读取更早内容。
func (s *service) ReadLog(id int64, after int64, before int64, pageSize int) (map[string]interface{}, error) {
	s.logLock.RLock()
	defer s.logLock.RUnlock()
	return cursorLog.Read(s.getTaskLogFilePath(id), after, before, pageSize)
}

// ClearLog 清理任务日志。
func (s *service) ClearLog(id int64) error {
	s.logLock.Lock()
	defer s.logLock.Unlock()
	return cursorLog.Clear(s.getTaskLogFile(id))
}

// WriteLog 写入日志
func (s *service) WriteLog(id int64, content string) error {
	if id < 0 {
		s.taskLock.Lock()
		defer s.taskLock.Unlock()
		if s.internalTasks[id] == nil {
			return nil
		}
	}
	s.logLock.Lock()
	defer s.logLock.Unlock()

	// 统一清理调用方携带的行结束符，确保每条记录只追加一个换行。
	content = strings.TrimRight(content, "\r\n")
	if content == "" {
		return nil
	}

	fileName := s.getTaskLogFilePath(id)
	if _, err := os.Stat(fileName); id >= 0 && os.IsNotExist(err) {
		if task, findErr := s.repositoryTask.FindById(id); findErr != nil || task == nil {
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

// validateCron 判断cron表达式
func (s *service) validateCron(expr string) bool {
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
