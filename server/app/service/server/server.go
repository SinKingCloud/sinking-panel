package server

import (
	"server/app/model"
	repositoryServer "server/app/repository/server"
	configService "server/app/service/config"
	"server/app/util/page"
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
	configService    configService.Service
}

// NewService 实例化service
func NewService(repositoryServer repositoryServer.Interface, config configService.Service) *service {
	return &service{
		repositoryServer: repositoryServer,
		configService:    config,
	}
}
