package container

// startRuntime 和 stopRuntime 将实例生命周期交给当前平台实现。
func (m *Manager) startRuntime(instance *Instance, image *Image, options RunOptions) error {
	return m.startPlatformRuntime(instance, image, options)
}

func (m *Manager) stopRuntime(instance *Instance) error {
	return m.stopPlatformRuntime(instance)
}

// prepareRuntimeRootfs 在启动前准备实例文件系统。
func (m *Manager) prepareRuntimeRootfs(instance *Instance, image *Image, writable bool) error {
	return m.preparePlatformRootfs(instance, image, writable)
}

// cleanupRuntimeRootfs 卸载实例的 merged rootfs。
func (m *Manager) cleanupRuntimeRootfs(instance *Instance) error {
	return m.cleanupPlatformRootfs(instance)
}

// execPlatform 在现有容器内执行命令。
func (m *Manager) execPlatform(id string, options ExecOptions) (*ExecResult, error) {
	return m.execPlatformCommand(id, options)
}
