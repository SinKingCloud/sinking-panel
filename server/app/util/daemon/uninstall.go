package daemon

import (
	"errors"
	"fmt"
	"os"
)

// uninstall 停止服务并移除系统自启动。
func (u *Daemon) uninstall() error {
	if err := u.Stop(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("停止服务失败: %w", err)
	}
	if err := u.UninstallAutoStart(); err != nil {
		return fmt.Errorf("卸载自启动失败: %w", err)
	}
	return nil
}
