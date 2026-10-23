package database

import (
	"server/app/model"
	utilDatabase "server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 数据库实例仓储接口。
type Interface interface {
	Create(data *model.Database, tx ...*gorm.DB) error
	DeleteByIds(ids []int64, tx ...*gorm.DB) error
	FindById(id int64, tx ...*gorm.DB) (*model.Database, error)
	Select(where *SelectDatabase, queryPage *page.Query) (*page.Result[*Database], error)
	UpdateById(id int64, data *UpdateDatabase, tx ...*gorm.DB) error
	ClearTypeId(typeIds []int64, tx ...*gorm.DB) error
}

// Repository 数据库实例仓储。
type Repository struct {
	*repository.Repository[*Database]
}

// NewRepository 创建数据库实例仓储。
func NewRepository(db *utilDatabase.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Database](db)}
}
