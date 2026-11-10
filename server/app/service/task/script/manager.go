package script

// Manager 校验并执行系统脚本，不依赖计划任务记录。
type Manager struct{}

// NewManager 创建系统脚本执行器。
func NewManager() *Manager {
	return &Manager{}
}
