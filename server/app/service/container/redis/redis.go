package redis

import (
	"sync"

	"server/app/util/container"
)

// Manager 管理指定 Redis 容器及其配置和数据。
type Manager struct {
	InstanceID       string
	container        *container.Manager
	root             string // redis.conf 保存目标配置，config/redis.conf 供下次启动读取，data 保存数据库文件。
	initErr          error
	mu               sync.Mutex // 串行处理配置和状态读取。
	runningConfig    []byte     // 当前运行代次的配置，保存新配置后仍使用原连接参数。
	configGeneration uint64
}

// NewManager 从容器管理器取得数据卷路径，不创建目录。
func NewManager(manager *container.Manager, instanceID string) *Manager {
	m := &Manager{container: manager, InstanceID: instanceID}
	m.root, m.initErr = manager.VolumePath(instanceID)
	return m
}
