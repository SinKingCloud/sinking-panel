package pgsql

import (
	"errors"
	"sync"

	serviceContainer "server/app/service/container"
)

// Manager 管理 PostgreSQL 容器及其配置和数据。
type Manager struct {
	InstanceID string
	container  serviceContainer.Service

	root             string
	initErr          error
	mu               sync.Mutex // 配置和账户操作串行，生命周期由通用容器管理器处理。
	runningConfig    []byte     // 当前进程已加载的配置，与保存后供下次启动读取的文件分开。
	configGeneration uint64
}

// NewManager 不创建目录；只有安装时才准备数据库文件。
func NewManager(manager serviceContainer.Service, instanceID string) *Manager {
	m := &Manager{container: manager, InstanceID: instanceID}
	m.root, m.initErr = manager.VolumePath(instanceID)

	return m
}

func (m *Manager) available() error {
	if m.initErr != nil {
		return m.initErr
	}
	if m.container == nil {
		return errors.New("容器管理器尚未初始化")
	}
	return nil
}
