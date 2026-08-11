package script

import "server/app/model"

// FindById 查询常用脚本详情
func (r *Repository) FindById(id int64) (data *model.Script, err error) {
	err = r.Database.Db.Where("`id` = ?", id).First(&data).Error
	return
}
