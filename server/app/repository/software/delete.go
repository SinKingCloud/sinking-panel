package software

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteById 删除已完成资源清理的软件记录。
func (r *Repository) DeleteById(id int64, tx ...*gorm.DB) error {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	result := db.Where("id = ?", id).Delete(&model.Software{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
