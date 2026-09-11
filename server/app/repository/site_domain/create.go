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
	return r.db(tx...).CreateInBatches(data, 500).Error
}
