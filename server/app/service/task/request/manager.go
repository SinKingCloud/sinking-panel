package request

// Manager 校验并执行 HTTP 请求，不依赖计划任务记录。
type Manager struct{}

// NewManager 创建 HTTP 请求执行器。
func NewManager() *Manager {
	return &Manager{}
}
