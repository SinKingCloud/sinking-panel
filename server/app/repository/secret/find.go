package secret

import (
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询密钥凭据详情。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (data *model.Secret, err error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err = db.Where("id = ?", id).First(&data).Error
	return
}
