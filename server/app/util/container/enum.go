package container

// Status 表示容器实例运行状态。
type Status string

const (
	StatusStarting Status = "starting" // 正在启动
	StatusRunning  Status = "running"  // 正在运行
	StatusStopping Status = "stopping" // 正在停止
	StatusStopped  Status = "stopped"  // 已停止
	StatusFailed   Status = "failed"   // 操作失败
	StatusUnknown  Status = "unknown"  // 平台状态暂时不可用
)
