package daemon

// AutoStartOptions 描述程序启动参数和系统自启动项。
// Executable、WorkingDirectory 默认使用当前程序和工作目录。
// Arguments 为 nil 时沿用当前进程参数，空切片表示不传参数；Name 仅管理系统自启动时必填。
type AutoStartOptions struct {
	Name               string
	Executable         string
	WorkingDirectory   string
	Arguments          []string
	KeepChildProcesses bool // Linux 服务退出时保留子进程，由应用决定如何清理。
}
