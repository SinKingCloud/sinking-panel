package server

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 服务器仓储接口
type Interface interface {
	Create(data *model.Server) error
	DeleteByIds(ids []int64, tx ...*gorm.DB) error
	FindById(id int64) (*model.Server, error)
	Select(where *SelectServer, queryPage *page.Query) (*page.Result[*Server], error)
	UpdateByIds(ids []int64, data *UpdateServer) error
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*Server]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Server](db)}
}
