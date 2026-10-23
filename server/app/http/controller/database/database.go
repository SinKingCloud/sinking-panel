package database

import (
	"server/app/http/controller/database/elasticsearch"
	"server/app/http/controller/database/mongo"
	"server/app/http/controller/database/mysql"
	"server/app/http/controller/database/postgres"
	"server/app/http/controller/database/redis"
	"server/app/http/controller/database/sqlite"
	"server/app/http/controller/database/sqlserver"
)

// 数据管理控制器，按实例及引擎统一访问。
var (
	Mysql         = &mysql.Controller{}
	Postgres      = &postgres.Controller{}
	Redis         = &redis.Controller{}
	Mongo         = &mongo.Controller{}
	SqlServer     = &sqlserver.Controller{}
	Sqlite        = &sqlite.Controller{}
	Elasticsearch = &elasticsearch.Controller{}
)
