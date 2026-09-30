package server

import (
	"server/app/model"
	repositoryServer "server/app/repository/server"
	serviceConfig "server/app/service/config"
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
	configService    serviceConfig.Service
	mu               sync.RWMutex // 配置和凭据的查询、组装、更新共用锁。
}

// NewService 实例化service
func NewService(repositoryServer repositoryServer.Interface, configService serviceConfig.Service) *service {
	return &service{
		repositoryServer: repositoryServer,
		configService:    configService,
	}
}
