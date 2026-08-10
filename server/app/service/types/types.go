package types

import (
	"server/app/model"
	repositoryScript "server/app/repository/script"
	repositoryTypes "server/app/repository/types"
	"server/app/util/cache"
	"server/app/util/database"
	"server/app/util/page"
)

// Service service接口
type Service interface {
	Create(data *model.Type) error
	DeleteByIds(ids []int64) error
	GetEnum(module string) (map[int64]string, error)
	Select(where *repositoryTypes.SelectType, queryPage *page.Query) (*page.Result[*model.Type], error)
	UpdateByIds(ids []int64, data *repositoryTypes.UpdateType) error
}

// service 注入结构
type service struct {
	database         *database.Database
	repositoryTypes  repositoryTypes.Interface
	repositoryScript repositoryScript.Interface
	cache            cache.Interface
}

// NewService 实例化service
func NewService(repositoryTypes repositoryTypes.Interface, repositoryScript repositoryScript.Interface, database *database.Database, cache cache.Interface) *service {
	return &service{
		database:         database,
		repositoryTypes:  repositoryTypes,
		repositoryScript: repositoryScript,
		cache:            cache,
	}
}
