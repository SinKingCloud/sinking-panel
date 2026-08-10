package service

import (
	configRepository "server/app/repository/config"
	logRepository "server/app/repository/log"
	serverRepository "server/app/repository/server"
	taskRepository "server/app/repository/task"
	"server/app/service/auth"
	"server/app/service/config"
	"server/app/service/file"
	logService "server/app/service/log"
	"server/app/service/recycle"
	serverService "server/app/service/server"
	"server/app/service/system"
	taskService "server/app/service/task"
	"server/global"
)

// service实例
var (
	Config  config.Service
	Auth    auth.Service
	Server  serverService.Service
	Log     logService.Service
	Task    taskService.Service
	File    file.Service
	Recycle recycle.Service
	System  system.Service
)

// Init 初始化service
func Init() {
	configRepo := configRepository.NewRepository(global.App.Database)
	logRepo := logRepository.NewRepository(global.App.Database)
	serverRepo := serverRepository.NewRepository(global.App.Database)
	taskRepo := taskRepository.NewRepository(global.App.Database)

	Config = config.NewService(configRepo, global.App.Cache)
	Auth = auth.NewService(Config, global.App.Cache)
	Server = serverService.NewService(serverRepo, Config)
	Log = logService.NewService(logRepo)
	Task = taskService.NewService(taskRepo)
	File = file.NewService()
	Recycle = recycle.NewService()
	System = system.NewService(File)
}
