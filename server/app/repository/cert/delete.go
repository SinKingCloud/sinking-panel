package cert

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteById 删除证书。
func (r *Repository) DeleteById(id int64, tx ...*gorm.DB) error {
	result := r.db(tx...).Where("id = ?", id).Delete(&model.Cert{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
