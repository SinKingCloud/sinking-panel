package app

import (
	"server/app/http/route"
	"server/app/service"
	"server/app/task"
)

func Run() {
	service.Init()
	task.Init()
	route.Init()
}
