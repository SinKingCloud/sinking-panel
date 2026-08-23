package site

import (
	"server/app/model"

	"gorm.io/gorm"
)

// Create 创建网站。
func (r *Repository) Create(data *model.Site, tx ...*gorm.DB) error {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(data).Error
}
