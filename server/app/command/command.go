package command

import (
	"server/app/command/queue"
	"server/app/command/task"
)

func Init() {
	task.Init()
	queue.Init()
}
