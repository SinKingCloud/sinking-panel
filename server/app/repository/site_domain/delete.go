package site_domain

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteBySiteId 删除网站的全部域名。
func (r *Repository) DeleteBySiteId(siteId int64, tx ...*gorm.DB) error {
	return r.db(tx...).Where("site_id = ?", siteId).Delete(&model.SiteDomain{}).Error
}
