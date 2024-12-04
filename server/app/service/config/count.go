package config

import (
	"server/app/util"
)

// CountByKey KEY数量
func (s *Service) CountByKey(key string) int64 {
	var total int64
	err := util.Database.Db.Where("`key` = ?", key).Count(&total).Error
	if err != nil {
		return 0
	}
	return total
}
