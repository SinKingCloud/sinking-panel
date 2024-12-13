package config

import (
	"server/app/model"
	"server/app/util"
)

// CountByKey KEY数量
func (s *Service) CountByKey(key string) (int64, error) {
	var total int64
	err := util.Database.Db.Model(&model.Config{}).Where("`key` = ?", key).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}
