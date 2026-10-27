package database

import (
	"context"
	"server/app/model"
	repositoryDatabase "server/app/repository/database"
	repositoryScript "server/app/repository/script"
	repositoryTypes "server/app/repository/types"
	clientElasticsearch "server/app/service/database/elasticsearch"
	clientMongo "server/app/service/database/mongo"
	clientMySQL "server/app/service/database/mysql"
	clientPostgres "server/app/service/database/postgres"
	clientRedis "server/app/service/database/redis"
	clientSqlite "server/app/service/database/sqlite"
	clientSqlServer "server/app/service/database/sqlserver"
	serviceSystem "server/app/service/system"
	serviceTypes "server/app/service/types"
	"server/app/util/cache"
	utilDatabase "server/app/util/database"
	"server/app/util/page"
)

// Service 管理实例连接，直接读取和操作实例中的数据库。
type Service interface {
	Config(databaseType int, content string, id int64, stored string, action string) (interface{}, error)
	Create(data *model.Database) error
	Update(id int64, data *repositoryDatabase.UpdateDatabase) error
	DeleteByIds(ids []int64) error
	FindById(id int64) (*Database, error)
	Select(where *repositoryDatabase.SelectDatabase, query *page.Query) (*page.Result[*repositoryDatabase.Database], error)
	FindConnection(id int64, databaseType int) (*model.Database, error)
	Test(ctx context.Context, id int64, databaseType int, config string) (map[string]interface{}, error)
	Transfer(id int64, databaseType int, action, path, databaseName string, run func(context.Context, func(float64, string)) (map[string]interface{}, error)) (string, error)
	Mysql(ctx context.Context, id int64, databaseName string) (clientMySQL.Interface, error)
	ChangeMysqlPassword(ctx context.Context, id int64, databaseName, name, host, password string) (interface{}, error)
	Postgres(ctx context.Context, id int64, databaseName string) (clientPostgres.Interface, error)
	ChangePostgresPassword(ctx context.Context, id int64, databaseName, name, password string) (interface{}, error)
	Sqlserver(ctx context.Context, id int64, databaseName string) (clientSqlServer.Interface, error)
	ChangeSqlServerPassword(ctx context.Context, id int64, databaseName, name, password string) (interface{}, error)
	Sqlite(ctx context.Context, id int64, databaseName string) (clientSqlite.Interface, error)
	Redis(ctx context.Context, id int64, databaseName string) (clientRedis.Interface, error)
	Mongo(ctx context.Context, id int64, databaseName string) (clientMongo.Interface, error)
	ChangeMongoPassword(ctx context.Context, id int64, databaseName, name, password string) (interface{}, error)
	Elasticsearch(ctx context.Context, id int64, databaseName string) (clientElasticsearch.Interface, error)
	RefreshElasticsearchFields(id int64, databaseName string) (string, error)
}

type service struct {
	repository       repositoryDatabase.Interface
	repositoryScript repositoryScript.Interface
	repositoryTypes  repositoryTypes.Interface
	typeService      serviceTypes.Service
	systemService    serviceSystem.Service
	database         *utilDatabase.Database
	cache            cache.Interface
}

// NewService 创建数据库客户端服务。
func NewService(repository repositoryDatabase.Interface, repositoryScript repositoryScript.Interface, repositoryTypes repositoryTypes.Interface, typeService serviceTypes.Service, systemService serviceSystem.Service, database *utilDatabase.Database, cache cache.Interface) *service {
	return &service{
		repository:       repository,
		repositoryScript: repositoryScript,
		repositoryTypes:  repositoryTypes,
		typeService:      typeService,
		systemService:    systemService,
		database:         database,
		cache:            cache,
	}
}
