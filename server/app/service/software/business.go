package software

import (
	"context"
	"errors"
	"log"
	"time"

	"server/app/enum/software_state"
	"server/app/enum/software_status"
	"server/app/enum/system_task_status"
	"server/app/model"
	repositorySoftware "server/app/repository/software"
	"server/app/util/page"
)

// entry 保存软件后台任务状态，插件加载和调用由 plugin.Manager 管理。
func (s *service) entry(id int64) (*entry, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, errors.New("软件管理正在关闭")
	}
	current := s.entries[id]
	if current == nil {
		current = &entry{}
		s.entries[id] = current
	}
	s.mu.Unlock()
	if current.busy.Load() && current.mu.TryLock() {
		err := s.restoreTask(current)
		current.mu.Unlock()
		if err != nil {
			return nil, err
		}
	}
	return current, nil
}

// restoreTask 仅恢复尚未进入回调的任务；调用方持有 entry.mu，执行中的任务不会被提前解锁。
func (s *service) restoreTask(e *entry) error {
	if e.pending == nil {
		return nil
	}
	if task := s.system.GetTask(e.taskID); task != nil && (task.Status == system_task_status.Pending || task.Status == system_task_status.Running) {
		return nil
	}
	record := e.pending
	var err error
	if e.created {
		err = s.repository.DeleteById(record.Id)
	} else {
		err = s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{State: &record.State})
	}
	s.invalidate(record.Code)
	s.invalidateManifest(record.Code)
	if err != nil {
		return err
	}
	e.pending = nil
	e.busy.Store(false)
	return nil
}

// Select 查询列表摘要；图标路径来自清单缓存，避免读取源码、配置和容器状态。
func (s *service) Select(where *repositorySoftware.SelectSoftware, query *page.Query) (*page.Result[*repositorySoftware.Software], error) {
	if where != nil && where.State != nil {
		if _, ok := software_state.Map()[*where.State]; !ok {
			return nil, errors.New("安装状态不合法")
		}
	}
	if where != nil && where.Status != nil {
		if _, ok := software_status.Map()[*where.Status]; !ok {
			return nil, errors.New("软件状态不合法")
		}
	}
	// 被取消的排队任务不会调用任务回调，读取列表时同步恢复其原状态。
	s.mu.Lock()
	entries := make([]*entry, 0, len(s.entries))
	for _, e := range s.entries {
		if e.busy.Load() {
			entries = append(entries, e)
		}
	}
	s.mu.Unlock()
	for _, e := range entries {
		if e.mu.TryLock() {
			err := s.restoreTask(e)
			e.mu.Unlock()
			if err != nil {
				return nil, err
			}
		}
	}
	result, err := s.repository.Select(where, query)
	if err != nil {
		return nil, err
	}
	for _, software := range result.List {
		s.mu.Lock()
		icon, cached := s.icons[software.Code]
		s.mu.Unlock()
		if !cached && s.plugins != nil {
			// 未 Boot 的服务也能返回图标；损坏的清单按无图标处理，不影响软件列表。
			if manifest, err := s.plugins.Manifest(software.Code); err == nil {
				icon = manifest.Icon
			}
			s.mu.Lock()
			if current, exists := s.icons[software.Code]; exists {
				// 并发安装可能已经更新缓存，保留已提交的新版本路径。
				icon = current
			} else {
				s.icons[software.Code] = icon
			}
			s.mu.Unlock()
		}
		software.Icon = icon
	}
	return result, nil
}

// Info 读取软件公开信息及页面入口，不执行插件代码或返回内部配置。
func (s *service) Info(id int64) (*Info, error) {
	record, err := s.repository.FindById(id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.New("软件不存在")
	}
	pkg, err := s.plugins.Info(record.Code)
	if err != nil {
		return nil, err
	}
	return &Info{
		Software: repositorySoftware.Software{
			Id:         record.Id,
			Name:       record.Name,
			Code:       record.Code,
			Version:    record.Version,
			Icon:       pkg.Manifest.Icon,
			Status:     record.Status,
			State:      record.State,
			CreateTime: record.CreateTime,
			UpdateTime: record.UpdateTime,
		},
		Frontend: pkg.Manifest.Frontend,
	}, nil
}

// Boot 按依赖顺序启动插件；失败保留安装状态，之后调用或启动面板时可重试。
func (s *service) Boot() {
	s.importing.Lock()
	s.mu.Lock()
	if s.closed || !s.booted.CompareAndSwap(false, true) {
		s.mu.Unlock()
		s.importing.Unlock()
		return
	}
	s.mu.Unlock()
	records, err := s.repository.SelectAll()
	if err != nil {
		log.Printf("读取插件启动列表失败: %v", err)
		s.importing.Unlock()
		return
	}
	// 新安装中断移除未提交代码；升级中断恢复备份，避免将半安装的软件启动。
	for _, record := range records {
		switch record.State {
		case software_state.Installing:
			if err := s.plugins.Remove(record.Code); err != nil {
				log.Printf("清理中断安装 %s 失败: %v", record.Code, err)
				continue
			}
			if err := s.repository.DeleteById(record.Id); err != nil {
				log.Printf("清理中断安装记录失败: %v", err)
			}
			s.invalidate(record.Code)
			continue
		case software_state.Upgrading:
			// 升级钩子可能已改写持久配置，异常退出恢复代码后先禁用，避免自动运行不一致版本。
			record.Status = software_status.Disabled
			backup := ".backup-" + record.Code
			if _, err := s.plugins.Manifest(backup); err == nil {
				if err = s.plugins.Remove(record.Code); err == nil {
					err = s.plugins.Rename(backup, record.Code)
				}
				if err != nil {
					log.Printf("恢复升级 %s 失败: %v", record.Code, err)
					continue
				}
			}
		case software_state.Removing:
		default:
			continue
		}
		state := software_state.Installed
		if err := s.repository.UpdateById(record.Id, &repositorySoftware.UpdateSoftware{State: &state, Status: &record.Status}); err != nil {
			log.Printf("恢复软件 %s 状态失败: %v", record.Code, err)
		} else {
			if record.State == software_state.Upgrading {
				log.Printf("软件 %s 升级被中断，已恢复并禁用，请检查配置后重新启用", record.Code)
			}
			record.State = state
			s.invalidate(record.Code)
		}
	}

	s.importing.Unlock()
	for _, record := range records {
		if record.Status != software_status.Enabled || record.State != software_state.Installed {
			continue
		}
		pkg, err := s.plugins.Info(record.Code)
		if err == nil {
			s.mu.Lock()
			s.icons[record.Code] = pkg.Manifest.Icon
			s.mu.Unlock()
		}
		if err == nil && !pkg.Available {
			err = errors.New(pkg.UnavailableReason)
		}
		if err == nil {
			var dependencies []*model.Software
			dependencies, err = s.dependencies(pkg.Manifest)
			if err == nil {
				var release func()
				release, err = s.start(append(dependencies, record))
				if release != nil {
					release()
				}
			}
		}
		if err != nil {
			log.Printf("启动插件 %s 失败: %v", record.Code, err)
		}
	}
}

// Close 先停止依赖方，再停止被依赖的软件，最后释放解释器，保留容器和数据。
func (s *service) Close() {
	// 安装提交或回滚完成后再关闭 Manager，确保临时目录完成清理。
	s.importing.Lock()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.importing.Unlock()
		return
	}
	s.closed = true
	entries := s.entries
	taskIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.busy.Load() {
			taskIDs = append(taskIDs, e.taskID)
		}
	}
	s.mu.Unlock()
	s.importing.Unlock()
	for _, taskID := range taskIDs {
		s.system.TaskCancel(taskID)
	}
	for _, e := range entries {
		e.mu.Lock()
		if e.pending != nil {
			if err := s.restoreTask(e); err != nil {
				log.Printf("恢复软件任务失败: %v", err)
			}
		}
		e.mu.Unlock()
	}
	if s.repository != nil && s.plugins != nil {
		records, err := s.repository.SelectAll()
		if err != nil {
			log.Printf("读取插件停止列表失败: %v", err)
		}
		byCode := make(map[string]*model.Software, len(records))
		for _, record := range records {
			byCode[record.Code] = record
		}
		ordered := make([]string, 0, len(records))
		visited := make(map[string]bool, len(records))
		var visit func(string)
		visit = func(code string) {
			if visited[code] || byCode[code] == nil {
				return
			}
			visited[code] = true
			// 依赖由包清单定义；清单损坏也继续停止，避免遗留后台任务。
			if manifest, err := s.plugins.Manifest(code); err == nil {
				for _, dependency := range manifest.Requires {
					visit(dependency)
				}
			}
			ordered = append(ordered, code)
		}
		for _, record := range records {
			visit(record.Code)
		}
		for index := len(ordered) - 1; index >= 0; index-- {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_, err := s.plugins.Lifecycle(ctx, ordered[index], "stop", "{}")
			cancel()
			if err != nil {
				log.Printf("停止插件 %s 失败: %v", ordered[index], err)
			}
		}
	}
	if s.plugins != nil {
		s.plugins.Close()
	}
}

// Page 读取 web 目录内的相对路径，省略路径时返回清单中的前端入口。
func (s *service) Page(code, path string) ([]byte, string, string, error) {
	record, err := s.record(code)
	if err != nil {
		return nil, "", "", err
	}
	if record == nil {
		return nil, "", "", errors.New("软件不存在")
	}
	id := record.Id
	e, err := s.entry(id)
	if err != nil {
		return nil, "", "", err
	}
	if !e.mu.TryRLock() {
		return nil, "", "", errors.New("软件正在执行操作")
	}
	defer e.mu.RUnlock()
	record, err = s.record(code)
	if err != nil {
		return nil, "", "", err
	}
	if record == nil || record.Id != id {
		return nil, "", "", errors.New("软件已变更，请重新打开")
	}
	if e.busy.Load() || record.State != software_state.Installed {
		return nil, "", "", errors.New("软件尚未就绪")
	}
	if record.Status != software_status.Enabled {
		return nil, "", "", errors.New("软件已禁用")
	}
	if path == "" {
		pkg, err := s.plugins.Info(record.Code)
		if err != nil {
			return nil, "", "", err
		}
		path = pkg.Manifest.Frontend
		if path == "" {
			return nil, "", "", errors.New("软件未配置页面入口")
		}
	} else {
		path = "web/" + path
	}
	return s.plugins.Page(record.Code, path)
}
