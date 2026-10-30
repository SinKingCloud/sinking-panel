package software

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"server/app/enum/software_state"
	"server/app/enum/software_status"
	"server/app/model"
	"server/app/util/plugin"
)

// dependencies 检查整条依赖链，返回依赖优先的安装记录，不包含当前软件。
// 升级以传入的新清单为根，因此增加依赖形成环时不会先停止旧版本。
func (s *service) dependencies(manifest *plugin.Manifest) ([]*model.Software, error) {
	ordered := make([]*model.Software, 0)
	visited := map[string]int{manifest.Id: 1}
	path := []string{manifest.Id}
	var visit func(string) error
	visit = func(code string) error {
		if visited[code] == 1 {
			return fmt.Errorf("软件存在循环依赖: %s", strings.Join(append(path, code), " → "))
		}
		if visited[code] == 2 {
			return nil
		}
		record, err := s.record(code)
		if err != nil {
			return err
		}
		if record == nil {
			return fmt.Errorf("缺少依赖软件 %s，请先安装", code)
		}
		if record.Status != software_status.Enabled || record.State != software_state.Installed {
			return fmt.Errorf("依赖软件 %s 尚未就绪或正在卸载", code)
		}
		pkg, err := s.plugins.Info(record.Code)
		if err != nil {
			return fmt.Errorf("依赖软件 %s 不可用: %w", code, err)
		}
		if pkg.Manifest.Id != code {
			return fmt.Errorf("依赖软件 %s 的标识与安装记录不一致", code)
		}
		if !pkg.Available {
			return fmt.Errorf("依赖软件 %s 不可用: %s", code, pkg.UnavailableReason)
		}
		visited[code] = 1
		path = append(path, code)
		for _, dependency := range pkg.Manifest.Requires {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		path = path[:len(path)-1]
		visited[code] = 2
		ordered = append(ordered, record)
		return nil
	}
	for _, code := range manifest.Requires {
		if err := visit(code); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// start 锁定整条依赖链并启动，调用方完成操作后释放，禁止依赖在中途被禁用或卸载。
func (s *service) start(records []*model.Software) (func(), error) {
	locked := make([]*entry, 0, len(records))
	release := func() {
		for i := len(locked) - 1; i >= 0; i-- {
			locked[i].mu.RUnlock()
		}
	}
	for _, record := range records {
		e, err := s.entry(record.Id)
		if err != nil {
			release()
			return nil, err
		}
		if !e.mu.TryRLock() {
			release()
			return nil, fmt.Errorf("软件 %s 正在执行操作", record.Code)
		}
		locked = append(locked, e)
		current, err := s.record(record.Code)
		if err != nil {
			release()
			return nil, err
		}
		if current == nil || current.Id != record.Id || e.busy.Load() || current.Status != software_status.Enabled || current.State != software_state.Installed {
			release()
			return nil, errors.New("软件尚未就绪或已禁用")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err = s.plugins.Lifecycle(ctx, current.Code, "start", "{}")
		cancel()
		if err != nil {
			release()
			return nil, fmt.Errorf("启动软件 %s 失败: %w", record.Code, err)
		}
	}
	return release, nil
}
