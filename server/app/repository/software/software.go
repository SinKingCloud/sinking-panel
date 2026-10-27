package software

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 软件仓储接口。
type Interface interface {
	Create(data *model.Software, tx ...*gorm.DB) error
	FindById(id int64, tx ...*gorm.DB) (*model.Software, error)
	FindByCode(code string, tx ...*gorm.DB) (*model.Software, error)
	Select(where *SelectSoftware, queryPage *page.Query) (*page.Result[*Software], error)
	SelectAll() ([]*model.Software, error)
	UpdateById(id int64, data *UpdateSoftware, tx ...*gorm.DB) error
	DeleteById(id int64, tx ...*gorm.DB) error
}

// Repository 软件仓储。
type Repository struct {
	*repository.Repository[*Software]
}

// NewRepository 创建软件仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Software](db)}
}
