package oss

import (
	"context"
	"server/app/model"
	repositoryOss "server/app/repository/oss"
	repositorySecret "server/app/repository/secret"
	serviceSystem "server/app/service/system"
	"server/app/util/database"
	utilOss "server/app/util/oss"
	"server/app/util/page"
	"sync"
)

// Service 管理 OSS 连接，文件操作复用统一对象存储客户端。
type Service interface {
	Create(data *model.Oss) error
	Update(id int64, data *repositoryOss.UpdateOss) error
	DeleteByIds(ids []int64) error
	FindById(id int64) (*Oss, error)
	Select(where *repositoryOss.SelectOss, query *page.Query) (*page.Result[*repositoryOss.Oss], error)
	Test(ctx context.Context, secretId int64, config string) error
	Client(id int64) (utilOss.Interface, error)
	Files(ctx context.Context, id int64, prefix, delimiter string, query *page.Query) (*page.Result[*File], error)
	Upload(id int64, key, path string, overwrite bool, options utilOss.PutOptions) (string, error)
	Download(id int64, key, path string, overwrite bool) (string, error)
	Copy(ctx context.Context, id int64, key, destination string, options utilOss.CopyOptions) (*utilOss.File, error)
	Move(ctx context.Context, id int64, key, destination string) (*utilOss.File, error)
	Delete(ctx context.Context, id int64, keys []string) ([]utilOss.DeleteResult, error)
}

// service 注入存储和密钥仓储，凭据始终从密钥管理读取。
type service struct {
	repository       repositoryOss.Interface
	repositorySecret repositorySecret.Interface
	systemService    serviceSystem.Service
	database         *database.Database
	operationMu      sync.Mutex
}

// NewService 创建对象存储管理服务。
func NewService(repository repositoryOss.Interface, repositorySecret repositorySecret.Interface, systemService serviceSystem.Service, database *database.Database) *service {
	return &service{
		repository:       repository,
		repositorySecret: repositorySecret,
		systemService:    systemService,
		database:         database,
	}
}
