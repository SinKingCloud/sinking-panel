package secret

import (
	"server/app/model"

	"gorm.io/gorm"
)

// Create 创建密钥凭据。
func (r *Repository) Create(data *model.Secret, tx ...*gorm.DB) error {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Create(data).Error
	}
	return r.Database.Db.Create(data).Error
}
