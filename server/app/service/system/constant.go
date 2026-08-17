package system

import "server/app/constant"

const (
	maxTaskWorkers         = 5                           // 系统任务并发 worker 数量
	systemTaskLogDirectory = constant.TempPath + "/task" // 系统任务日志目录
)
