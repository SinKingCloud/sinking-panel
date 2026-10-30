package software

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"server/app/enum/software_state"
	"server/app/enum/system_task_status"
	repositorySoftware "server/app/repository/software"
	"server/app/util/str"
)

// Uninstall 按标识或 ID 卸载软件并清理插件数据，提交后台任务后立即返回任务 ID。
func (s *service) Uninstall(target string) (int64, error) {
	target = strings.TrimSpace(target)
	record, err := s.repository.FindByCode(target)
	if err != nil {
		return 0, err
	}
	var id int64
	if record != nil {
		id = record.Id
	} else {
		id, err = strconv.ParseInt(target, 10, 64)
		if err != nil || id <= 0 {
			return 0, errors.New("未找到该软件，请检查软件标识或 ID")
		}
	}
	// 提交阶段持有软件锁；排队任务也会取得此锁，因此等待前必须释放。
	taskID, err := func() (int64, error) {
		// 与安装的依赖检查串行，提交卸载状态后即可释放，不占用后台任务执行时间。
		if !s.importing.TryLock() {
			return 0, errors.New("软件正在安装或卸载，请稍后重试")
		}
		defer s.importing.Unlock()
		e, err := s.entry(id)
		if err != nil {
			return 0, err
		}
		if !e.mu.TryLock() {
			return 0, errors.New("软件任务正在执行")
		}
		defer e.mu.Unlock()
		if e.busy.Load() {
			return 0, errors.New("软件任务正在执行")
		}
		record, err := s.repository.FindById(id)
		if err != nil {
			return 0, err
		}
		if record == nil {
			return 0, errors.New("软件不存在")
		}
		if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`).MatchString(record.Code) {
			return 0, errors.New("软件标识无效")
		}
		records, err := s.repository.SelectAll()
		if err != nil {
			return 0, err
		}
		var dependents []string
		for _, other := range records {
			if other.Id == id {
				continue
			}
			manifest, err := s.manifest(other.Code)
			if err != nil {
				return 0, fmt.Errorf("读取软件 %s 的依赖失败: %w", other.Code, err)
			}
			if slices.Contains(manifest.Requires, record.Code) {
				dependents = append(dependents, other.Name+"（"+other.Code+"）")
			}
		}
		if len(dependents) > 0 {
			return 0, fmt.Errorf("软件仍被 %s 依赖，请先卸载这些软件或更新其依赖", strings.Join(dependents, "、"))
		}
		// 卸载按标识清理固定目录，不依赖软件配置或代码能否加载。
		directory := record.Code
		// 重试中断的卸载，失败或取消时恢复为已安装状态。
		if record.State == software_state.Removing {
			record.State = software_state.Installed
		}
		taskID := str.GetSnowWorkIns().GetId()
		key := strconv.FormatInt(taskID, 10)
		state := software_state.Removing
		if err = s.repository.UpdateById(id, &repositorySoftware.UpdateSoftware{State: &state}); err != nil {
			return 0, err
		}
		s.invalidate(record.Code)
		e.created = false
		e.busy.Store(true)
		e.pending = record
		e.taskID = key
		s.system.TaskCreate(key, "卸载 "+record.Name, map[string]interface{}{"software_id": id}, func(ctx context.Context, _ interface{}, progress func(int, float64, string)) {
			e.mu.Lock()
			defer e.mu.Unlock()
			if e.taskID != key {
				return
			}
			if ctx.Err() != nil {
				s.system.TaskCancel(key)
				if err := s.restoreTask(e); err != nil {
					progress(system_task_status.Failed, 100, err.Error())
				}
				return
			}
			e.pending = nil
			defer e.busy.Store(false)
			defer s.invalidate(record.Code)
			defer s.invalidateManifest(record.Code)
			progress(system_task_status.Running, 10, "正在卸载软件")
			ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
			_, runErr := s.plugins.Lifecycle(ctx, directory, "stop", "{}")
			if runErr == nil {
				_, runErr = s.plugins.Lifecycle(ctx, directory, "uninstall", `{"remove_data":true}`)
			}
			if runErr == nil {
				// 清理失败时保留记录，方便再次卸载。
				runErr = s.plugins.Remove(directory)
			}
			if runErr == nil {
				runErr = s.repository.DeleteById(id)
			}
			if runErr != nil {
				saveErr := s.repository.UpdateById(id, &repositorySoftware.UpdateSoftware{State: &record.State})
				message := errors.Join(runErr, saveErr).Error()
				progress(system_task_status.Failed, 100, message)
				return
			}
			s.mu.Lock()
			delete(s.icons, record.Code)
			s.mu.Unlock()
			progress(system_task_status.Completed, 100, "卸载完成")
		})
		if s.system.GetTask(key) == nil {
			return 0, errors.Join(errors.New("创建后台任务失败，面板正在停止"), s.restoreTask(e))
		}
		return taskID, nil
	}()
	if err != nil {
		return 0, err
	}
	return taskID, nil
}
