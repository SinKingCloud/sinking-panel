package server

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 查询信息
func (r *Repository) FindById(id int64) (*model.Server, error) {
	var data *model.Server
	err := r.Database.Db.
		Model(&model.Server{}).
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
