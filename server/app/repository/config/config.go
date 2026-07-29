package config

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/repository"
)

// Interface 配置仓储接口
type Interface interface {
	FindByGroup(group string) ([]*model.Config, error)
	CountByKey(key string) (int64, error)
	UpdateByKey(key string, value string) error
	Create(config *model.Config) error
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*model.Config]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*model.Config](db)}
}
