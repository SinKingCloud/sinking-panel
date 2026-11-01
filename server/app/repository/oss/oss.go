package oss

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 对象存储连接仓储接口。
type Interface interface {
	Create(data *model.Oss, tx ...*gorm.DB) error
	UpdateById(id int64, data *UpdateOss, tx ...*gorm.DB) error
	DeleteByIds(ids []int64, tx ...*gorm.DB) error
	FindById(id int64, tx ...*gorm.DB) (*model.Oss, error)
	CountBySecretId(id int64, tx ...*gorm.DB) (int64, error)
	Select(where *SelectOss, query *page.Query) (*page.Result[*Oss], error)
}

// Repository 对象存储连接仓储。
type Repository struct {
	*repository.Repository[*Oss]
}

// NewRepository 创建对象存储连接仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Oss](db)}
}
