package server

import "server/app/model"

// DeleteByIds 通过ID列表删除
func (r *Repository) DeleteByIds(ids []int64) error {
	return r.Database.Db.Where("`id` in ? ", ids).Delete(&model.Server{}).Error
}
