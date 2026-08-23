package cert

import (
	"server/app/model"

	"gorm.io/gorm"
)

// Create 创建证书。
func (r *Repository) Create(data *model.Cert, tx ...*gorm.DB) error {
	return r.db(tx...).Create(data).Error
}
