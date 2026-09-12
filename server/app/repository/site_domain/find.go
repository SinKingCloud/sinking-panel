package site_domain

import (
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询网站域名。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (data *model.SiteDomain, err error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err = db.Where("id = ?", id).First(&data).Error
	return
}

// Exists 查询域名是否已被其他网站使用。
func (r *Repository) Exists(domain string, excludeSiteId int64, tx ...*gorm.DB) (bool, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	query := db.Model(&model.SiteDomain{}).Where("LOWER(domain) = LOWER(?)", domain)
	if excludeSiteId > 0 {
		query = query.Where("site_id <> ?", excludeSiteId)
	}
	var count int64
	err := query.Limit(1).Count(&count).Error
	return count > 0, err
}
