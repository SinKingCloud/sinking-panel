package secret

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateById 更新密钥凭据。
func (r *Repository) UpdateById(id int64, data *UpdateSecret, tx ...*gorm.DB) error {
	if data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.Provider != nil {
		updates["provider"] = *data.Provider
	}
	if data.Data != nil {
		updates["data"] = *data.Data
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	execute := func(db *gorm.DB) error {
		result := db.
			Model(&model.Secret{}).
			Where("id = ?", id).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return execute(r.Database.Db)
}
