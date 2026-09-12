package service

import (
	repositoryCert "server/app/repository/cert"
	repositoryConfig "server/app/repository/config"
	repositoryLog "server/app/repository/log"
	repositoryScript "server/app/repository/script"
	repositorySecret "server/app/repository/secret"
	repositoryServer "server/app/repository/server"
	repositorySite "server/app/repository/site"
	repositorySiteDomain "server/app/repository/site_domain"
	repositoryTask "server/app/repository/task"
	repositoryTypes "server/app/repository/types"
	serviceAuth "server/app/service/auth"
	serviceConfig "server/app/service/config"
	serviceFile "server/app/service/file"
	serviceLog "server/app/service/log"
	serviceRecycle "server/app/service/recycle"
	serviceScript "server/app/service/script"
	serviceSecret "server/app/service/secret"
	serviceServer "server/app/service/server"
	serviceSite "server/app/service/site"
	serviceSystem "server/app/service/system"
	serviceTask "server/app/service/task"
	serviceTypes "server/app/service/types"
	"server/global"
)

// service实例
var (
	Config  serviceConfig.Service
	Auth    serviceAuth.Service
	Server  serviceServer.Service
	Script  serviceScript.Service
	Secret  serviceSecret.Service
	Type    serviceTypes.Service
	Log     serviceLog.Service
	Task    serviceTask.Service
	File    serviceFile.Service
	Recycle serviceRecycle.Service
	System  serviceSystem.Service
	Site    serviceSite.Service
)

// Init 初始化service
func Init() {
	configRepository := repositoryConfig.NewRepository(global.App.Database)
	logRepository := repositoryLog.NewRepository(global.App.Database)
	serverRepository := repositoryServer.NewRepository(global.App.Database)
	taskRepository := repositoryTask.NewRepository(global.App.Database)
	typesRepository := repositoryTypes.NewRepository(global.App.Database)
	scriptRepository := repositoryScript.NewRepository(global.App.Database)
	siteRepository := repositorySite.NewRepository(global.App.Database)
	siteDomainRepository := repositorySiteDomain.NewRepository(global.App.Database)
	certRepository := repositoryCert.NewRepository(global.App.Database)
	secretRepository := repositorySecret.NewRepository(global.App.Database)

	Config = serviceConfig.NewService(configRepository, global.App.Cache)
	Auth = serviceAuth.NewService(Config, global.App.Cache)
	Server = serviceServer.NewService(serverRepository, Config)
	Script = serviceScript.NewService(scriptRepository)
	Secret = serviceSecret.NewService(secretRepository, certRepository, global.App.Database, global.App.Cache)
	Type = serviceTypes.NewService(typesRepository, scriptRepository, siteRepository, taskRepository, global.App.Database, global.App.Cache)
	Log = serviceLog.NewService(logRepository)
	Task = serviceTask.NewService(taskRepository, Type)
	File = serviceFile.NewService(global.App.Cache)
	Recycle = serviceRecycle.NewService()
	System = serviceSystem.NewService(File)
	var err error
	Site, err = serviceSite.NewService(siteRepository, siteDomainRepository, certRepository, secretRepository, Type, Config, global.App.Database, global.App.Cache)
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
