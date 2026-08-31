package types

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 类型仓储接口
type Interface interface {
	Create(data *model.Type) error
	DeleteByIds(ids []int64, tx ...*gorm.DB) error
	SelectByIds(ids []int64, tx ...*gorm.DB) ([]*model.Type, error)
	SelectIdNameMap(module string) (map[int64]string, error)
	Select(where *SelectType, orderByField, orderByType string) ([]*model.Type, error)
	UpdateByIds(ids []int64, data *UpdateType, tx ...*gorm.DB) error
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*model.Type]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*model.Type](db)}
}
