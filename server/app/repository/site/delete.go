package site

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteById 删除网站。
func (r *Repository) DeleteById(id int64, tx ...*gorm.DB) error {
	execute := func(db *gorm.DB) error {
		result := db.Where("id = ?", id).Delete(&model.Site{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return execute(r.Database.Db)
}
