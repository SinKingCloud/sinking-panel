package oss

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateById 更新指定连接的信息。
func (r *Repository) UpdateById(id int64, data *UpdateOss, tx ...*gorm.DB) error {
	if data == nil {
		return nil
	}
	values := make(map[string]interface{})
	if data.Name != nil {
		values["name"] = *data.Name
	}
	if data.Config != nil {
		values["config"] = *data.Config
	}
	if len(values) == 0 {
		return nil
	}
	values["update_time"] = str.DateTime(time.Now())
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	result := db.Model(&model.Oss{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
