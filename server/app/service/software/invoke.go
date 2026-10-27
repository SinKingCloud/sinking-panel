package software

import (
	"errors"
	"strings"

	"server/app/enum/software_status"
	"server/app/util/context"
)

// Invoke 校验软件状态；插件的方法分发、并发控制和错误处理交给 Manager。
func (s *service) Invoke(c *context.Context, code, method string) error {
	if c == nil || c.Context == nil || c.Request == nil {
		return errors.New("软件请求上下文不能为空")
	}
	method = strings.TrimSpace(method)
	record, err := s.repository.FindByCode(code)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("软件不存在")
	}
	id := record.Id
	e, err := s.entry(id)
	if err != nil {
		return err
	}
	// 保持本次调用所用的软件状态和版本，更新、卸载不能跨过正在执行的调用。
	if !e.mu.TryRLock() {
		return errors.New("软件任务正在执行")
	}
	defer e.mu.RUnlock()
	if e.busy.Load() {
		return errors.New("软件任务正在执行")
	}
	record, err = s.repository.FindById(id)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("软件不存在")
	}
	if record.Status != software_status.Installed {
		return errors.New("软件任务正在执行")
	}
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
	if err = s.start(append(dependencies, record)); err != nil {
		return err
	}
	return s.plugins.Call(c.Request.Context(), record.Code, method, c)
}
