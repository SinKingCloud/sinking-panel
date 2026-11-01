package secret

import (
	"server/app/model"
	repositoryCert "server/app/repository/cert"
	repositoryOss "server/app/repository/oss"
	repositorySecret "server/app/repository/secret"
	"server/app/util/database"
	"server/app/util/page"
	"sync"
)

// Service 密钥管理服务。
type Service interface {
	Create(data *model.Secret) error
	Update(id int64, data *repositorySecret.UpdateSecret) error
	Delete(id int64) error
	FindById(id int64) (*Secret, error)
	Select(where *repositorySecret.SelectSecret, queryPage *page.Query) (*page.Result[*repositorySecret.Secret], error)
}

type service struct {
	repositorySecret repositorySecret.Interface
	repositoryCert   repositoryCert.Interface
	repositoryOss    repositoryOss.Interface
	database         *database.Database
	operationMu      sync.Mutex
}

// NewService 创建密钥管理服务。
func NewService(repositorySecret repositorySecret.Interface, repositoryCert repositoryCert.Interface, repositoryOss repositoryOss.Interface, database *database.Database) *service {
	return &service{
		repositorySecret: repositorySecret,
		repositoryCert:   repositoryCert,
		repositoryOss:    repositoryOss,
		database:         database,
	}
}
