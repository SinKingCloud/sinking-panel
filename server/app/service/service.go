package service

import (
	"server/app/constant"
	repositoryCert "server/app/repository/cert"
	repositoryConfig "server/app/repository/config"
	repositoryDatabase "server/app/repository/database"
	repositoryLog "server/app/repository/log"
	repositoryOss "server/app/repository/oss"
	repositoryScript "server/app/repository/script"
	repositorySecret "server/app/repository/secret"
	repositoryServer "server/app/repository/server"
	repositorySite "server/app/repository/site"
	repositorySiteBackup "server/app/repository/site_backup"
	repositorySiteDomain "server/app/repository/site_domain"
	repositorySoftware "server/app/repository/software"
	repositoryTask "server/app/repository/task"
	repositoryTypes "server/app/repository/types"
	serviceAuth "server/app/service/auth"
	serviceConfig "server/app/service/config"
	serviceContainer "server/app/service/container"
	serviceDatabase "server/app/service/database"
	serviceFile "server/app/service/file"
	serviceLog "server/app/service/log"
	serviceOss "server/app/service/oss"
	serviceRecycle "server/app/service/recycle"
	serviceScript "server/app/service/script"
	serviceSecret "server/app/service/secret"
	serviceServer "server/app/service/server"
	serviceSite "server/app/service/site"
	serviceSoftware "server/app/service/software"
	"server/app/service/software/sdk"
	serviceSystem "server/app/service/system"
	serviceTask "server/app/service/task"
	serviceTypes "server/app/service/types"
	"server/global"
)

// service实例
var (
	Config    serviceConfig.Service
	Auth      serviceAuth.Service
	Server    serviceServer.Service
	Script    serviceScript.Service
	Secret    serviceSecret.Service
	Type      serviceTypes.Service
	Log       serviceLog.Service
	Task      serviceTask.Service
	Oss       serviceOss.Service
	File      serviceFile.Service
	Recycle   serviceRecycle.Service
	System    serviceSystem.Service
	Site      serviceSite.Service
	Container serviceContainer.Service
	Software  serviceSoftware.Service
	Database  serviceDatabase.Service
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
	siteBackupRepository := repositorySiteBackup.NewRepository(global.App.Database)
	siteDomainRepository := repositorySiteDomain.NewRepository(global.App.Database)
	certRepository := repositoryCert.NewRepository(global.App.Database)
	secretRepository := repositorySecret.NewRepository(global.App.Database)
	ossRepository := repositoryOss.NewRepository(global.App.Database)
	databaseRepository := repositoryDatabase.NewRepository(global.App.Database)

	Config = serviceConfig.NewService(configRepository, global.App.Cache)
	Auth = serviceAuth.NewService(Config, global.App.Cache)
	Server = serviceServer.NewService(serverRepository, scriptRepository, typesRepository, Config, global.App.Database, global.App.Cache)
	Secret = serviceSecret.NewService(secretRepository, certRepository, ossRepository, global.App.Database)
	Type = serviceTypes.NewService(typesRepository, scriptRepository, siteRepository, taskRepository, databaseRepository, global.App.Database, global.App.Cache)
	Script = serviceScript.NewService(scriptRepository, typesRepository)
	Log = serviceLog.NewService(logRepository)
	System = serviceSystem.NewService(Config, global.App.Daemon, global.App.Container, global.App.Cache)
	Oss = serviceOss.NewService(ossRepository, secretRepository, System, global.App.Database)
	Task = serviceTask.NewService(taskRepository, siteRepository, siteBackupRepository, Type, Oss)
	Recycle = serviceRecycle.NewService()
	File = serviceFile.NewService(Recycle, System, global.App.Cache)
	Database = serviceDatabase.NewService(databaseRepository, scriptRepository, typesRepository, Type, System, global.App.Database, global.App.Cache)
	Container = serviceContainer.NewManager(global.App.Container)
	Site = nil
	if site, err := serviceSite.NewService(siteRepository, siteDomainRepository, certRepository, secretRepository, siteBackupRepository, Type, Config, Task, Recycle, Oss, File, System, global.App.Database, global.App.Cache); err == nil {
		Site = site
	}
	services := &sdk.Service{
		Config:    Config,
		Auth:      Auth,
		Server:    Server,
		Script:    Script,
		Secret:    Secret,
		Type:      Type,
		Log:       Log,
		Task:      Task,
		File:      File,
		Recycle:   Recycle,
		System:    System,
		Site:      Site,
		Container: Container,
		Database:  Database,
	}
	Software = serviceSoftware.NewService(repositorySoftware.NewRepository(global.App.Database), System, sdk.NewManager(constant.SoftwarePath, sdk.App{
		Service:  services,
		Database: global.App.Database,
		Cache:    global.App.Cache,
		Config:   global.App.Config,
		Log:      global.App.Log,
	}))
	services.Software = Software
}

// Boot 启动服务运行时，并返回统一的清理函数。
func Boot() func() {
	if System != nil {
		System.Start()
	}
	if Site != nil {
		Site.Boot()
	}
	if Task != nil {
		Task.Start()
	}
	if Software != nil {
		Software.Boot()
	}
	return func() {
		// 插件先停止，清理钩子仍可使用计划任务、网站和系统服务。
		if Software != nil {
			Software.Close()
		}
		if Site != nil {
			Site.Close()
		}
		if Task != nil {
			Task.Close()
		}
		if File != nil {
			File.Close()
		}
		if Database != nil {
			Database.Close()
		}
		if Oss != nil {
			Oss.Close()
		}
		if System != nil {
			System.Close()
		}
	}
}
