package task

import "server/app/service"

// Init 初始化计划任务。
func Init() {
	service.Task.Start()
}
