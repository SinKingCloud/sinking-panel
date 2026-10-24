package server

import (
	"server/app/model"
	repositoryScript "server/app/repository/script"
	repositoryServer "server/app/repository/server"
	repositoryTypes "server/app/repository/types"
	serviceConfig "server/app/service/config"
	"server/app/util/cache"
	"server/app/util/database"
	"server/app/util/page"
	"sync"
)

// Service service接口
type Service interface {
	Create(data *model.Server) error
	DeleteByIds(ids []int64) error
	FindById(id int64) (*model.Server, error)
	Select(where *repositoryServer.SelectServer, queryPage *page.Query) (*page.Result[*repositoryServer.Server], error)
	UpdateByIds(ids []int64, data *repositoryServer.UpdateServer) error
}

// service 注入结构
type service struct {
	repositoryServer repositoryServer.Interface
	repositoryScript repositoryScript.Interface
	repositoryTypes  repositoryTypes.Interface
	configService    serviceConfig.Service
	database         *database.Database
	cache            cache.Interface
	mu               sync.RWMutex // 配置和凭据的查询、组装、更新共用锁。
}

// NewService 实例化service
func NewService(repositoryServer repositoryServer.Interface, repositoryScript repositoryScript.Interface, repositoryTypes repositoryTypes.Interface, configService serviceConfig.Service, database *database.Database, cache cache.Interface) *service {
	return &service{
		repositoryServer: repositoryServer,
		repositoryScript: repositoryScript,
		repositoryTypes:  repositoryTypes,
		configService:    configService,
		database:         database,
		cache:            cache,
	}
}
