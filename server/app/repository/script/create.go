package script

import "server/app/model"

// Create 创建常用脚本
func (r *Repository) Create(data *model.Script) error {
	return r.Database.Db.Create(data).Error
}
