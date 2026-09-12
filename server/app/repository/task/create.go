package task

import "server/app/model"

// Create 创建任务
func (r *Repository) Create(data *model.Task) error {
	return r.Database.Db.Create(data).Error
}
