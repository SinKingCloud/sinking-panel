package software

import (
	serviceContainer "server/app/service/container"
	serviceMongo "server/app/service/software/mongo"
	serviceMysql "server/app/service/software/mysql"
	servicePgsql "server/app/service/software/pgsql"
	serviceRedis "server/app/service/software/redis"
)

// Service 内置软件管理入口，容器和镜像操作交由注入的容器服务处理。
type Service interface {
	Redis() *serviceRedis.Manager // 获取 Redis 管理器
	Mysql() *serviceMysql.Manager // 获取 MySQL 管理器
	Pgsql() *servicePgsql.Manager // 获取 PostgreSQL 管理器
	Mongo() *serviceMongo.Manager // 获取 MongoDB 管理器
}

type service struct {
	redis *serviceRedis.Manager
	mysql *serviceMysql.Manager
	pgsql *servicePgsql.Manager
	mongo *serviceMongo.Manager
}

// NewService 将同一个容器服务注入各软件管理器。
func NewService(containerService serviceContainer.Service) *service {
	return &service{
		redis: serviceRedis.NewManager(containerService, "redis"),
		mysql: serviceMysql.NewManager(containerService, "mysql"),
		pgsql: servicePgsql.NewManager(containerService, "pgsql"),
		mongo: serviceMongo.NewManager(containerService, "mongo"),
	}
}
