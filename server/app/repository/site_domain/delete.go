package site_domain

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteBySiteId 删除网站的全部域名。
func (r *Repository) DeleteBySiteId(siteId int64, tx ...*gorm.DB) error {
	execute := func(db *gorm.DB) error {
		return db.Where("site_id = ?", siteId).Delete(&model.SiteDomain{}).Error
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return execute(r.Database.Db)
}
