package software

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"server/app/enum/software_state"
	"server/app/enum/software_status"
	repositorySoftware "server/app/repository/software"
)

// Update 修改名称和启用状态；同一软件与接口调用、安装、卸载互斥。
func (s *service) Update(id int64, data *repositorySoftware.UpdateSoftware) error {
	if data == nil {
		return nil
	}
	changes := &repositorySoftware.UpdateSoftware{Name: data.Name, Status: data.Status}
	if changes.Name != nil {
		name := strings.TrimSpace(*changes.Name)
		if name == "" || len([]rune(name)) > 100 {
			return errors.New("软件名称不能为空且不能超过100个字符")
		}
		changes.Name = &name
	}
	if changes.Status != nil {
		if _, ok := software_status.Map()[*changes.Status]; !ok {
			return errors.New("软件启用状态不合法")
		}
	}
	e, err := s.entry(id)
	if err != nil {
		return err
	}
	if !e.mu.TryLock() {
		return errors.New("软件正在执行操作，请稍后重试")
	}
	defer e.mu.Unlock()
	if e.busy.Load() {
		return errors.New("软件任务正在执行")
	}
	record, err := s.repository.FindById(id)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("软件不存在")
	}
	if record.State != software_state.Installed {
		return errors.New("软件尚未安装完成")
	}
	changed := changes.Status != nil && *changes.Status != record.Status
	if changed {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if *changes.Status == software_status.Enabled {
			pkg, err := s.plugins.Info(record.Code)
			if err != nil {
				return err
			}
			if !pkg.Available {
				return errors.New(pkg.UnavailableReason)
			}
			dependencies, err := s.dependencies(pkg.Manifest)
			if err != nil {
				return err
			}
			release, err := s.start(dependencies)
			if err != nil {
				return err
			}
			defer release()
			if _, err = s.plugins.Lifecycle(ctx, record.Code, "start", "{}"); err != nil {
				return err
			}
		} else {
			records, err := s.repository.SelectAll()
			if err != nil {
				return err
			}
			for _, other := range records {
				if other.Id == id || other.Status != software_status.Enabled {
					continue
				}
				manifest, err := s.manifest(other.Code)
				if err != nil {
					return fmt.Errorf("读取软件 %s 的依赖失败: %w", other.Code, err)
				}
				if slices.Contains(manifest.Requires, record.Code) {
					return fmt.Errorf("软件仍被 %s 依赖，请先禁用该软件", other.Name)
				}
			}
			// 先写禁用状态，停止失败也禁止后续请求重新启动插件。
			if err = s.repository.UpdateById(id, changes); err != nil {
				return err
			}
			s.invalidate(record.Code)
			_, err = s.plugins.Lifecycle(ctx, record.Code, "stop", "{}")
			return err
		}
	}
	err = s.repository.UpdateById(id, changes)
	s.invalidate(record.Code)
	if err != nil && changed && *changes.Status == software_status.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, stopErr := s.plugins.Lifecycle(ctx, record.Code, "stop", "{}")
		err = errors.Join(err, stopErr)
	}
	// 允许重试禁用，完成上一次停止失败的清理。
	if err == nil && changes.Status != nil && *changes.Status == software_status.Disabled {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err = s.plugins.Lifecycle(ctx, record.Code, "stop", "{}")
	}
	return err
}
