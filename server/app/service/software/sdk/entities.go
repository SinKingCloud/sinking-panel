// Package sdk 定义软件包与宿主之间的稳定调用边界。
package sdk

import (
	"log"
	serviceAuth "server/app/service/auth"
	serviceConfig "server/app/service/config"
	serviceContainer "server/app/service/container"
	serviceDatabase "server/app/service/database"
	serviceFile "server/app/service/file"
	serviceLog "server/app/service/log"
	serviceRecycle "server/app/service/recycle"
	serviceScript "server/app/service/script"
	serviceSecret "server/app/service/secret"
	serviceServer "server/app/service/server"
	serviceSite "server/app/service/site"
	serviceSoftware "server/app/service/software"
	serviceSystem "server/app/service/system"
	serviceTask "server/app/service/task"
	serviceTypes "server/app/service/types"
	"server/app/util/cache"
	"server/app/util/database"

	"github.com/spf13/viper"
)

// App 提供当前软件的实例标识和面板基础能力。
type App struct {
	Service     *Service           // 面板已有业务服务，初始化和关闭由面板管理
	Database    *database.Database // 面板数据库，共用连接池、事务和批量操作
	Cache       cache.Interface    // 面板缓存，插件使用独立键前缀
	Config      *viper.Viper       // 面板运行配置
	Log         *log.Logger        // 面板日志
	InstanceID  string
	PackagePath string // 当前软件包的绝对路径，供读取包内镜像等离线资源
}

// Service 汇集面板已有服务，插件直接复用实例，不重复创建连接或运行时。
type Service struct {
	Config    serviceConfig.Service
	Auth      serviceAuth.Service
	Server    serviceServer.Service
	Script    serviceScript.Service
	Secret    serviceSecret.Service
	Type      serviceTypes.Service
	Log       serviceLog.Service
	Task      serviceTask.Service
	File      serviceFile.Service
	Recycle   serviceRecycle.Service
	System    serviceSystem.Service
	Site      serviceSite.Service
	Container serviceContainer.Service
	Software  serviceSoftware.Service
	Database  serviceDatabase.Service
}
