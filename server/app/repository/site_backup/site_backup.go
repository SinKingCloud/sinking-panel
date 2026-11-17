package site_backup

import (
	"server/app/model"
	"server/app/util/database"
	"server/app/util/page"
	"server/app/util/repository"
)

// Interface 网站备份仓储接口。
type Interface interface {
	Create(data *model.SiteBackup) error
	FindById(id int64) (*model.SiteBackup, error)
	Select(where *SelectSiteBackup, queryPage *page.Query) (*page.Result[*SiteBackup], error)
	SelectBySiteId(id int64) ([]*model.SiteBackup, error)
	SelectByIds(ids []int64) ([]*model.SiteBackup, error)
	UpdateData(id int64, before, data string) (int64, error)
	DeleteByIds(ids []int64) error
}

// Repository 网站备份仓储。
type Repository struct {
	*repository.Repository[*SiteBackup]
}

// NewRepository 创建网站备份仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Repository: repository.NewRepository[*SiteBackup](db)}
}
