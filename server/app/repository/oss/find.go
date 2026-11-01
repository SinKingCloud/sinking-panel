package oss

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 按 ID 获取对象存储连接，不存在时返回空值。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (*model.Oss, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var data model.Oss
	if err := db.Where("id = ?", id).Take(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}
