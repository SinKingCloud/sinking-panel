package config

import "server/app/model"

// CountByKey 获取KEY数量
func (r *Repository) CountByKey(key string) (int64, error) {
	var total int64
	err := r.Database.Db.Model(&model.Config{}).Where("`key` = ?", key).Count(&total).Error
	return total, err
}
