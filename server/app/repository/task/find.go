package task

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过ID查询任务
func (r *Repository) FindById(id int64) (*model.Task, error) {
	var data *model.Task
	err := r.Database.Db.
		Model(&model.Task{}).
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
