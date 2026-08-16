package log

import (
	"time"

	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"
)

// Interface 日志仓储接口
type Interface interface {
	Create(data *model.Log) error
	Select(where *SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error)
	SelectIdByCreateTime(before time.Time, limit int) ([]int64, error)
	Delete(ids []int64) error
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*model.Log]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*model.Log](db)}
}
