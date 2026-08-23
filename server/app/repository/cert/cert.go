package cert

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 证书仓储接口。
type Interface interface {
	Create(data *model.Cert, tx ...*gorm.DB) error
	DeleteById(id int64, tx ...*gorm.DB) error
	FindById(id int64, tx ...*gorm.DB) (*model.Cert, error)
	Select(where *SelectCert, queryPage *page.Query) (*page.Result[*Cert], error)
	SelectByIds(ids []int64, tx ...*gorm.DB) ([]*model.Cert, error)
	UpdateById(id int64, data *UpdateCert, tx ...*gorm.DB) error
}

// Repository 证书仓储。
type Repository struct {
	*repository.Repository[*Cert]
}

// NewRepository 创建证书仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Cert](db)}
}

func (r *Repository) db(tx ...*gorm.DB) *gorm.DB {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0]
	}
	return r.Database.Db
}
