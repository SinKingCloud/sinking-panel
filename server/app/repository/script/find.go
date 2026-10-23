package script

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 查询常用脚本详情
func (r *Repository) FindById(id int64) (*model.Script, error) {
	var data *model.Script
	err := r.Database.Db.
		Model(&model.Script{}).
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
