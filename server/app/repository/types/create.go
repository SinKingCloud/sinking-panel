package types

import (
	"server/app/model"

	"gorm.io/gorm"
)

// Create 插入数据
func (r *Repository) Create(data *model.Type) error {
	return r.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(data).Error; err != nil {
			return err
		}
		data.Sort = data.Id
		return tx.Model(data).UpdateColumn("sort", data.Sort).Error
	})
}
