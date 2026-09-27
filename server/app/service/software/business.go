package software

import (
	serviceMongo "server/app/service/software/mongo"
	serviceMysql "server/app/service/software/mysql"
	servicePgsql "server/app/service/software/pgsql"
	serviceRedis "server/app/service/software/redis"
)

// Redis 返回当前服务管理的 Redis 实例。
func (s *service) Redis() *serviceRedis.Manager {
	return s.redis
}

// Mysql 返回当前服务管理的 MySQL 实例。
func (s *service) Mysql() *serviceMysql.Manager {
	return s.mysql
}

// Pgsql 返回当前服务管理的 PostgreSQL 实例。
func (s *service) Pgsql() *servicePgsql.Manager {
	return s.pgsql
}

// Mongo 返回当前服务管理的 MongoDB 实例。
func (s *service) Mongo() *serviceMongo.Manager {
	return s.mongo
}
