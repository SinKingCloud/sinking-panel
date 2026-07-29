package task

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"
	"time"
)

// Interface 任务仓储接口
type Interface interface {
	Create(data *model.Task) error
	DeleteById(id int64) error
	FindById(id int64) (*model.Task, error)
	SelectAll() ([]*model.Task, error)
	Select(where *SelectTask, queryPage *page.Query) (*page.Result[*Task], error)
	UpdateEntryIDById(id int64, entryID int) error
	UpdateStateById(id int64, entryID int, status int) error
	UpdateRuntimeById(id int64, runTime time.Time) error
	UpdateByIds(ids []int64, data *UpdateTask) error
}

// Repository 仓储
type Repository struct {
	*repository.Repository[*Task]
}

// NewRepository 创建仓储
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*Task](db)}
}
