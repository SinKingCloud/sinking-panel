package secret

import (
	"server/app/model"
	repositoryCert "server/app/repository/cert"
	repositorySecret "server/app/repository/secret"
	"server/app/util/cache"
	"server/app/util/database"
	"server/app/util/page"
	"sync"
)

// Service 密钥管理服务。
type Service interface {
	Create(data *model.Secret) error
	Update(id int64, data *repositorySecret.UpdateSecret) error
	Delete(id int64) error
	FindById(id int64) (*model.Secret, error)
	Select(where *repositorySecret.SelectSecret, queryPage *page.Query) (*page.Result[*repositorySecret.Secret], error)
	GetIdNameMap(where *repositorySecret.SelectSecret) (map[int64]string, error)
}

type service struct {
	repositorySecret repositorySecret.Interface
	repositoryCert   repositoryCert.Interface
	database         *database.Database
	cache            cache.Interface
	enumMu           sync.Mutex
}

// NewService 创建密钥管理服务。
func NewService(repositorySecret repositorySecret.Interface, repositoryCert repositoryCert.Interface, database *database.Database, cache cache.Interface) *service {
	return &service{
		repositorySecret: repositorySecret,
		repositoryCert:   repositoryCert,
		database:         database,
		cache:            cache,
	}
}
