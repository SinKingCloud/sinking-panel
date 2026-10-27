package config

import "server/app/model"

// DeleteByGroup 删除指定配置组，不匹配名称相近的其他组。
func (r *Repository) DeleteByGroup(group string) error {
	return r.Database.Db.Where("`key` = ? OR (`key` >= ? AND `key` < ?)", group, group+".", group+"/").
		Delete(&model.Config{}).Error
}
