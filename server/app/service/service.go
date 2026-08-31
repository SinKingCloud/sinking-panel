package service

import (
	certRepository "server/app/repository/cert"
	configRepository "server/app/repository/config"
	domainRepository "server/app/repository/domain"
	logRepository "server/app/repository/log"
	scriptRepository "server/app/repository/script"
	serverRepository "server/app/repository/server"
	siteRepository "server/app/repository/site"
	taskRepository "server/app/repository/task"
	typeRepository "server/app/repository/types"
	"server/app/service/auth"
	"server/app/service/config"
	"server/app/service/file"
	logService "server/app/service/log"
	"server/app/service/recycle"
	scriptService "server/app/service/script"
	serverService "server/app/service/server"
	siteService "server/app/service/site"
	"server/app/service/system"
	taskService "server/app/service/task"
	typeService "server/app/service/types"
	"server/global"
)

// service实例
var (
	Config  config.Service
	Auth    auth.Service
	Server  serverService.Service
	Script  scriptService.Service
	Type    typeService.Service
	Log     logService.Service
	Task    taskService.Service
	File    file.Service
	Recycle recycle.Service
	System  system.Service
	Site    siteService.Service
)

// Init 初始化service
func Init() {
	configRepo := configRepository.NewRepository(global.App.Database)
	logRepo := logRepository.NewRepository(global.App.Database)
	serverRepo := serverRepository.NewRepository(global.App.Database)
	taskRepo := taskRepository.NewRepository(global.App.Database)
	typeRepo := typeRepository.NewRepository(global.App.Database)
	scriptRepo := scriptRepository.NewRepository(global.App.Database)
	siteRepo := siteRepository.NewRepository(global.App.Database)
	domainRepo := domainRepository.NewRepository(global.App.Database)
	certRepo := certRepository.NewRepository(global.App.Database)

	Config = config.NewService(configRepo, global.App.Cache)
	Auth = auth.NewService(Config, global.App.Cache)
	Server = serverService.NewService(serverRepo, Config)
	Script = scriptService.NewService(scriptRepo)
	Type = typeService.NewService(typeRepo, scriptRepo, siteRepo, taskRepo, global.App.Database, global.App.Cache)
	Log = logService.NewService(logRepo)
	Task = taskService.NewService(taskRepo, Type)
	File = file.NewService(global.App.Cache)
	Recycle = recycle.NewService()
	System = system.NewService(File)
	var err error
	Site, err = siteService.NewService(siteRepo, domainRepo, certRepo, Type, Config, global.App.Database, global.App.Cache)
	if err != nil {
		panic(err)
	}
}

// Boot 启动服务运行时，并返回统一的清理函数。
func Boot() func() {
	System.Start()
	Task.Start()
	stopSite := Site.Boot()
	return func() {
		stopSite()
		Task.Close()
		System.Close()
	}
}
