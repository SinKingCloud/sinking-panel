package task

// Status 任务状态
type Status int

const (
	StatusPending   Status = iota // 等待中
	StatusRunning                 // 运行中
	StatusCompleted               // 已完成
	StatusFailed                  // 失败
	StatusCanceled                // 已取消
)

// Status 获取所有任务状态
func (s *Service) Status() map[Status]string {
	return map[Status]string{
		StatusPending:   "等待中",
		StatusRunning:   "运行中",
		StatusCompleted: "已完成",
		StatusFailed:    "失败",
		StatusCanceled:  "已取消",
	}
}
