package domain

import (
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询网站域名。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (data *model.Domain, err error) {
	err = r.db(tx...).Where("id = ?", id).First(&data).Error
	return
}

// Exists 查询域名是否已被其他网站使用。
func (r *Repository) Exists(domain string, excludeSiteId int64, tx ...*gorm.DB) (bool, error) {
	query := r.db(tx...).Model(&model.Domain{}).Where("LOWER(domain) = LOWER(?)", domain)
	if excludeSiteId > 0 {
		query = query.Where("site_id <> ?", excludeSiteId)
	}
	var count int64
	err := query.Limit(1).Count(&count).Error
	return count > 0, err
}
