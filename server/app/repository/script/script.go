package script

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"

	"gorm.io/gorm"
)

// Interface 常用脚本仓储接口
type Interface interface {
	Create(data *model.Script) error
	DeleteByIds(ids []int64, tx ...*gorm.DB) error
	Select(where *SelectScript, queryPage *page.Query) (*page.Result[*model.Script], error)
	UpdateByIds(ids []int64, data *UpdateScript) error
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*model.Script]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*model.Script](db)}
}
