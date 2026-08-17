package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/enum/task_exec_type"
	"server/app/enum/task_status"
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/util/file"
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
	data, err := s.typeService.FindById(typeId)
	if err != nil || data == nil || data.Module != type_module.Task {
		return errors.New("任务分类不合法")
	}
	return nil
}

// checkContent 判断任务内容
func (s *service) checkContent(content string, execType int) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("任务内容不能为空")
	}
	switch execType {
	case task_exec_type.Script:
		return content, nil
	case task_exec_type.Request:
		data := Request{}
		if err := json.Unmarshal([]byte(content), &data); err != nil {
			data = Request{Method: http.MethodGet, Url: strings.TrimSpace(content)}
		}
		data.Method = strings.ToUpper(strings.TrimSpace(data.Method))
		if data.Method == "" {
			data.Method = http.MethodGet
		}
		switch data.Method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
		default:
			return "", errors.New("请求方式不合法")
		}
		data.Url = strings.TrimSpace(data.Url)
		request, err := http.NewRequest(data.Method, data.Url, nil)
		if err != nil || request.URL.Host == "" || (!strings.EqualFold(request.URL.Scheme, "http") && !strings.EqualFold(request.URL.Scheme, "https")) {
			return "", errors.New("请求地址必须是有效的HTTP或HTTPS地址")
		}
		if len(data.Headers) > 100 {
			return "", errors.New("请求头数量不能超过100个")
		}
		validHeaderName := func(name string) bool {
			for i := 0; i < len(name); i++ {
				char := name[i]
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(char)) {
					continue
				}
				return false
			}
			return name != ""
		}
		validHeaderValue := func(value string) bool {
			for i := 0; i < len(value); i++ {
				char := value[i]
				if char != '\t' && (char < 0x20 || char == 0x7f) {
					return false
				}
			}
			return true
		}
		headers := make(map[string]string, len(data.Headers))
		headerNames := make(map[string]struct{}, len(data.Headers))
		for name, value := range data.Headers {
			name = strings.TrimSpace(name)
			if !validHeaderName(name) {
				return "", errors.New("请求头名称不合法")
			}
			lowerName := strings.ToLower(name)
			if _, ok := headerNames[lowerName]; ok {
				return "", errors.New("请求头名称不能重复")
			}
			if lowerName == "content-length" || lowerName == "transfer-encoding" || lowerName == "trailer" {
				return "", errors.New("不支持自定义传输层请求头")
			}
			if !validHeaderValue(value) {
				return "", errors.New("请求头值不合法")
			}
			if lowerName == "host" && strings.TrimSpace(value) == "" {
				return "", errors.New("Host请求头不能为空")
			}
			headerNames[lowerName] = struct{}{}
			name = http.CanonicalHeaderKey(name)
			headers[name] = value
			if lowerName == "host" {
				request.Host = value
			} else {
				request.Header.Set(name, value)
			}
		}
		if err = request.Write(io.Discard); err != nil {
			return "", errors.New("请求配置不合法")
		}
		data.Headers = headers
		value, err := json.Marshal(data)
		if err != nil {
			return "", errors.New("请求配置格式化失败")
		}
		return string(value), nil
	default:
		return "", errors.New("任务类型不合法")
	}
}

// formatContent 格式化任务内容
func (s *service) formatContent(content string, execType int) interface{} {
	if execType == task_exec_type.Script {
		return content
	}
	if execType == task_exec_type.Request {
		content, err := s.checkContent(content, execType)
		if err == nil {
			data := Request{}
			if json.Unmarshal([]byte(content), &data) == nil {
				return data
			}
		}
		return Request{}
	}
	return nil
}

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
	content, err := s.checkContent(data.Script, data.ExecType)
	if err != nil {
		return err
	}
	data.Script = content
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

// ReadLog 读取日志。首次读取最新内容，after 读取新增内容，before 读取更早内容。
func (s *service) ReadLog(id int64, after int64, before int64, pageSize int) map[string]interface{} {
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
			"cursor":       after,
			"start_cursor": after,
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
			"cursor":       after,
			"start_cursor": after,
			"has_previous": false,
			"end":          false,
		}
	}
	result := map[string]interface{}{
		"file_name":    filepath.Base(fileName),
		"lines":        []string{},
		"cursor":       after,
		"start_cursor": after,
		"has_previous": after > 0,
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
