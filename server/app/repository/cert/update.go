package cert

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateById 更新证书。
func (r *Repository) UpdateById(id int64, data *UpdateCert, tx ...*gorm.DB) error {
	if data == nil {
		return nil
	}
	updates := map[string]interface{}{}
	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.Type != nil {
		updates["type"] = *data.Type
	}
	if data.Domains != nil {
		updates["domains"] = *data.Domains
	}
	if data.Certificate != nil {
		updates["certificate"] = *data.Certificate
	}
	if data.PrivateKey != nil {
		updates["private_key"] = *data.PrivateKey
	}
	if data.StartTime != nil {
		updates["start_time"] = *data.StartTime
	}
	if data.ExpireTime != nil {
		updates["expire_time"] = *data.ExpireTime
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	result := r.db(tx...).Model(&model.Cert{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
