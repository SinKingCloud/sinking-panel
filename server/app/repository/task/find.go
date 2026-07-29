package task

import "server/app/model"

// FindById 通过ID查询任务
func (r *Repository) FindById(id int64) (data *model.Task, err error) {
	err = r.Database.Db.Where("`id` = ? ", id).First(&data).Error
	return
}
