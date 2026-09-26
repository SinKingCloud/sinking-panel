package container_status

import "server/app/util/container"

// Map 容器运行状态数据。
func Map() map[container.State]string {
	return map[container.State]string{
		container.StatusStarting: "启动中",
		container.StatusRunning:  "运行中",
		container.StatusStopping: "停止中",
		container.StatusStopped:  "已停止",
		container.StatusFailed:   "运行失败",
		container.StatusUnknown:  "未知状态",
	}
}
