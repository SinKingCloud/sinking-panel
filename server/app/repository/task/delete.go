package task

import "server/app/model"

// DeleteById 通过ID删除任务
func (r *Repository) DeleteById(id int64) error {
	return r.Database.Db.Where("`id` = ? ", id).Delete(&model.Task{}).Error
}
