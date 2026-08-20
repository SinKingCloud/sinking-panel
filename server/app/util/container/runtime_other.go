//go:build !linux

package container

import (
	"errors"
	"os"
)

func (m *Manager) startPlatformRuntime(instance *Instance, image *Image, options RunOptions) error {
	return errors.New("libcontainer 仅支持 Linux")
}

func (m *Manager) stopPlatformRuntime(instance *Instance) error {
	return errors.New("libcontainer 仅支持 Linux")
}

func (m *Manager) preparePlatformRootfs(instance *Instance, image *Image, writable bool) error {
	return errors.New("libcontainer 仅支持 Linux")
}

func (m *Manager) cleanupPlatformRootfs(instance *Instance) error {
	return nil
}

func (m *Manager) cleanupOrphanPlatformRuntime(id string) error {
	return errPlatformRuntimeUnavailable
}

func (m *Manager) execPlatformCommand(id string, options ExecOptions) (*ExecResult, error) {
	return nil, errors.New("libcontainer 仅支持 Linux")
}

func (m *Manager) openPlatformTerminal(id string, height, width int) (TerminalSession, error) {
	return nil, errors.New("libcontainer 仅支持 Linux")
}

func (m *Manager) recoverPlatformRuntime(instance *Instance) (bool, error) {
	if m.isActiveStatus(instance.Status) || instance.PID > 0 {
		return false, errors.New("当前平台无法恢复 Linux 容器运行状态")
	}
	return false, nil
}

func (m *Manager) refreshPlatformRuntime(instance *Instance) (bool, error) {
	return false, errors.New("当前平台无法读取 Linux 容器运行状态")
}

func (m *Manager) statsPlatform(id string) (*Stats, error) {
	return nil, errors.New("libcontainer 仅支持 Linux")
}

func (m *Manager) clearPlatformLog(id, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	closeErr := file.Close()
	var removeErr error
	for _, name := range []string{path + ".1", path + ".1.tmp"} {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			removeErr = errors.Join(removeErr, err)
		}
	}
	return errors.Join(closeErr, removeErr)
}

func (m *Manager) lockPlatformLog(id string) func() {
	return func() {}
}

func (m *Manager) setPlatformFailure(id string, failure error) {}
