package log

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"
)

// Interface 日志仓储接口
type Interface interface {
	Create(data *model.Log) error
	Select(where *SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error)
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*model.Log]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*model.Log](db)}
}
