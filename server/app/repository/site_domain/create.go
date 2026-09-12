package site_domain

import (
	"server/app/model"

	"gorm.io/gorm"
)

// CreateBatch 批量创建网站域名。
func (r *Repository) CreateBatch(data []*model.SiteDomain, tx ...*gorm.DB) error {
	if len(data) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(data, 500, func(batch interface{}) error {
			return db.Create(batch.([]*model.SiteDomain)).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return execute(tx)
	})
}
