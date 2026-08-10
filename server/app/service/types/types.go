package types

import (
	"server/app/model"
	repositoryTypes "server/app/repository/types"
	"server/app/util/database"
	"server/app/util/page"
)

// Service service接口
type Service interface {
	Create(data *model.Type) error
	DeleteByIds(ids []int64) error
	Select(where *repositoryTypes.SelectType, queryPage *page.Query) (*page.Result[*model.Type], error)
	UpdateByIds(ids []int64, data *repositoryTypes.UpdateType) error
}

// service 注入结构
type service struct {
	database        *database.Database
	repositoryTypes repositoryTypes.Interface
}

// NewService 实例化service
func NewService(repositoryTypes repositoryTypes.Interface, database *database.Database) *service {
	return &service{
		database:        database,
		repositoryTypes: repositoryTypes,
	}
}
