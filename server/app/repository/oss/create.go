package oss

import (
	"server/app/model"

	"gorm.io/gorm"
)

// Create 保存对象存储连接信息。
func (r *Repository) Create(data *model.Oss, tx ...*gorm.DB) error {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(data).Error
}
