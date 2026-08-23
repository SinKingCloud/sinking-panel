package cert

import (
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 通过 ID 查询证书详情。
func (r *Repository) FindById(id int64, tx ...*gorm.DB) (data *model.Cert, err error) {
	err = r.db(tx...).Where("id = ?", id).First(&data).Error
	return
}
