package system

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // 等待中
	TaskStatusRunning                     // 运行中
	TaskStatusCompleted                   // 已完成
	TaskStatusFailed                      // 失败
	TaskStatusCanceled                    // 已取消
)

// TaskStatus 获取所有任务状态
func (s *Service) TaskStatus() map[TaskStatus]string {
	return map[TaskStatus]string{
		TaskStatusPending:   "等待中",
		TaskStatusRunning:   "运行中",
		TaskStatusCompleted: "已完成",
		TaskStatusFailed:    "失败",
		TaskStatusCanceled:  "已取消",
	}
}
