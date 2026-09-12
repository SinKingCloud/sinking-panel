package site_domain

import (
	"server/app/model"
	"server/app/util/database"

	"gorm.io/gorm"
)

// Interface 网站域名仓储接口。
type Interface interface {
	CreateBatch(data []*model.SiteDomain, tx ...*gorm.DB) error
	CountByCertId(certId int64, tx ...*gorm.DB) (int64, error)
	DeleteBySiteId(siteId int64, tx ...*gorm.DB) error
	Exists(domain string, excludeSiteId int64, tx ...*gorm.DB) (bool, error)
	FindById(id int64, tx ...*gorm.DB) (*model.SiteDomain, error)
	SelectBySiteId(siteId int64, tx ...*gorm.DB) ([]*model.SiteDomain, error)
	SelectBySiteIds(siteIds []int64, tx ...*gorm.DB) ([]*model.SiteDomain, error)
	UpdateCertId(id, certId int64, tx ...*gorm.DB) error
}

// Repository 网站域名仓储。
type Repository struct {
	Database *database.Database
}

// NewRepository 创建网站域名仓储。
func NewRepository(db *database.Database) *Repository {
	return &Repository{Database: db}
}
