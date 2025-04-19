package command

import "server/app/service"

func Init() {
	service.Task.Start()
}
