package config

import "server/app/model"

// FindByGroup 获取配置数据
func (r *Repository) FindByGroup(group string) ([]*model.Config, error) {
	var configs []*model.Config
	query := r.Database.Db.Model(&model.Config{})
	if group != "" {
		query = query.Where("`key` like ?", group+"%")
	}
	err := query.Find(&configs).Error
	return configs, err
}
