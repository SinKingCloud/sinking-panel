package secret

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询密钥凭据详情。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (*model.Secret, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var data *model.Secret
	err := db.
		Model(&model.Secret{}).
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
