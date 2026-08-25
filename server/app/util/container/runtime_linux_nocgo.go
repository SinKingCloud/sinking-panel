//go:build linux && !cgo

package container

import (
	"errors"
	"os"
	fileLog "server/app/util/log"
)

func (m *Manager) configurePlatformSecurity(root string) error {
	return errors.New("容器安全运行需要使用包含静态 libseccomp 的 Linux 发布构建")
}

func (m *Manager) normalizePlatformRootfsOwnership(root string) error {
	return nil
}

func (m *Manager) preparePlatformMountOwnership(root string) error {
	return errors.New("准备容器挂载需要启用 cgo")
}

func (m *Manager) validatePlatformMountOwnership(root string) error {
	return errors.New("验证容器挂载需要启用 cgo")
}

func (m *Manager) validatePlatformMountContent(root string) error {
	return errors.New("验证容器挂载需要启用 cgo")
}

func (m *Manager) startPlatformRuntime(instance *Instance, image *Image, options RunOptions) error {
	return errors.New("libcontainer 运行需要启用 cgo")
}

func (m *Manager) stopPlatformRuntime(instance *Instance) error {
	return errors.New("libcontainer 运行需要启用 cgo")
}

func (m *Manager) preparePlatformRootfs(instance *Instance, image *Image, writable bool) error {
	return errors.New("libcontainer 运行需要启用 cgo")
}

func (m *Manager) cleanupPlatformRootfs(instance *Instance) error {
	return nil
}

func (m *Manager) cleanupOrphanPlatformRuntime(id string) error {
	return errPlatformRuntimeUnavailable
}

func (m *Manager) execPlatformCommand(id string, options ExecOptions) (*ExecResult, error) {
	return nil, errors.New("libcontainer 运行需要启用 cgo")
}

func (m *Manager) openPlatformTerminal(id string, height, width int) (TerminalSession, error) {
	return nil, errors.New("libcontainer 运行需要启用 cgo")
}

func (m *Manager) recoverPlatformRuntime(instance *Instance) (bool, error) {
	if m.isActiveStatus(instance.Status) || instance.PID > 0 {
		return false, errors.New("未启用 cgo，无法恢复容器运行状态")
	}
	return false, nil
}

func (m *Manager) refreshPlatformRuntime(instance *Instance) (bool, error) {
	return false, errors.New("未启用 cgo，无法读取容器运行状态")
}

func (m *Manager) statsPlatform(id string) (*Stats, error) {
	return nil, errors.New("libcontainer 运行需要启用 cgo")
}

func (m *Manager) clearPlatformLog(id, path string) error {
	clearErr := fileLog.Clear(path)
	var removeErr error
	for _, name := range []string{path + ".1", path + ".1.tmp"} {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			removeErr = errors.Join(removeErr, err)
		}
	}
	return errors.Join(clearErr, removeErr)
}

func (m *Manager) lockPlatformLog(id string) func() {
	return func() {}
}

func (m *Manager) setPlatformFailure(id string, failure error) {}
