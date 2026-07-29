package task_status

const (
	Running = iota //运行
	Stop           //暂停
)

// Map 任务状态数据
func Map() map[int]string {
	return map[int]string{
		Running: "运行",
		Stop:    "暂停",
	}
}
