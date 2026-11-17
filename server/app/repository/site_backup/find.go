package site_backup

import (
	"errors"
	"server/app/model"

	"gorm.io/gorm"
)

// FindById 根据 ID 查询网站备份。
func (r *Repository) FindById(id int64) (*model.SiteBackup, error) {
	var data *model.SiteBackup
	err := r.Database.Db.Model(&model.SiteBackup{}).Where("id = ?", id).First(&data).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if data == nil || data.Id <= 0 {
		return nil, nil
	}
	return data, nil
}
