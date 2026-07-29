package system_task_status

const (
	Pending   = iota //等待中
	Running          //运行中
	Completed        //已完成
	Failed           //失败
	Canceled         //已取消
)

// Map 系统任务状态数据
func Map() map[int]string {
	return map[int]string{
		Pending:   "等待中",
		Running:   "运行中",
		Completed: "已完成",
		Failed:    "失败",
		Canceled:  "已取消",
	}
}
