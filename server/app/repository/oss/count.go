package oss

import (
	"server/app/model"

	"gorm.io/gorm"
)

// CountBySecretId 统计引用指定密钥的对象存储连接。
func (r *Repository) CountBySecretId(id int64, tx ...*gorm.DB) (int64, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var count int64
	err := db.Model(&model.Oss{}).Where("secret_id = ?", id).Count(&count).Error
	return count, err
}
