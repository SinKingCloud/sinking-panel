package site

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 网站仓储接口。
type Interface interface {
	Create(data *model.Site, tx ...*gorm.DB) error
	DeleteById(id int64, tx ...*gorm.DB) error
	FindById(id int64, tx ...*gorm.DB) (*model.Site, error)
	Select(where *SelectSite, queryPage *page.Query) (*page.Result[*Site], error)
	SelectAll(tx ...*gorm.DB) ([]*model.Site, error)
	UpdateById(id int64, data *UpdateSite, tx ...*gorm.DB) error
}

// Repository 网站仓储。
type Repository struct {
	*repository.Repository[*Site]
}

// NewRepository 创建网站仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Site](db)}
}
