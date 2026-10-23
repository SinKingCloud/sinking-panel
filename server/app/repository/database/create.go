package database

import (
	"server/app/model"

	"gorm.io/gorm"
)

// Create 创建数据库实例。
func (r *Repository) Create(data *model.Database, tx ...*gorm.DB) error {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Create(data).Error
	}
	return r.Database.Db.Create(data).Error
}
