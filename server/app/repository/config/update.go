package config

import (
	"server/app/model"
	"server/app/util/str"
	"time"
)

// UpdateByKey 通过KEY更新配置
func (r *Repository) UpdateByKey(key string, value string) error {
	return r.Database.Db.Model(&model.Config{}).Where("`key` = ? ", key).Updates(map[string]any{
		"update_time": str.DateTime(time.Now()),
		"value":       value,
	}).Error
}
