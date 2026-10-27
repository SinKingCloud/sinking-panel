package software

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询软件。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (*model.Software, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var data *model.Software
	err := db.Model(&model.Software{}).Where("id = ?", id).First(&data).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if data == nil || data.Id <= 0 {
		return nil, nil
	}
	return data, nil
}

// FindByCode 通过唯一标识查询软件。
func (r *Repository) FindByCode(code string, tx ...*gorm.DB) (*model.Software, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var data *model.Software
	err := db.Model(&model.Software{}).Where("code = ?", code).First(&data).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if data == nil || data.Id <= 0 {
		return nil, nil
	}
	return data, nil
}
