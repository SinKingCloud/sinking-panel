package server

import "server/app/model"

// FindById 查询信息
func (r *Repository) FindById(id int64) (data *model.Server, err error) {
	err = r.Database.Db.Where("`id` = ?", id).First(&data).Error
	return
}
