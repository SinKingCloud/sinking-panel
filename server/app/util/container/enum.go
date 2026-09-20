package container

// State 表示容器实例运行状态。
type State string

const (
	StatusStarting State = "starting" // 正在启动
	StatusRunning  State = "running"  // 正在运行
	StatusStopping State = "stopping" // 正在停止
	StatusStopped  State = "stopped"  // 已停止
	StatusFailed   State = "failed"   // 操作失败
	StatusUnknown  State = "unknown"  // 平台状态暂时不可用
)
