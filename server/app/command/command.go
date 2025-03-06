package command

import "server/app/service"

func Init() {
	service.Cron.Start()
}
