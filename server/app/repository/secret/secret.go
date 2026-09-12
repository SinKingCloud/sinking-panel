package secret

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 密钥凭据仓储接口。
type Interface interface {
	Create(data *model.Secret, tx ...*gorm.DB) error
	DeleteById(id int64, tx ...*gorm.DB) error
	FindById(id int64, tx ...*gorm.DB) (*model.Secret, error)
	Select(where *SelectSecret, queryPage *page.Query) (*page.Result[*Secret], error)
	SelectIdNameMap(where *SelectSecret) (map[int64]string, error)
	UpdateById(id int64, data *UpdateSecret, tx ...*gorm.DB) error
}

// Repository 密钥凭据仓储。
type Repository struct {
	*repository.Repository[*Secret]
}

// NewRepository 创建密钥凭据仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Secret](db)}
}
