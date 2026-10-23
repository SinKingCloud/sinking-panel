package cert

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询证书详情。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (*model.Cert, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var data *model.Cert
	err := db.
		Model(&model.Cert{}).
		Where("id = ?", id).
		First(&data).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if data == nil || data.Id <= 0 {
		return nil, nil
	}
	return data, nil
}
