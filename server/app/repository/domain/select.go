package domain

import (
	"server/app/model"

	"gorm.io/gorm"
)

// SelectBySiteId 查询网站的全部域名。
func (r *Repository) SelectBySiteId(siteId int64, tx ...*gorm.DB) ([]*model.Domain, error) {
	return r.SelectBySiteIds([]int64{siteId}, tx...)
}

// SelectBySiteIds 查询多个网站的全部域名。
func (r *Repository) SelectBySiteIds(siteIds []int64, tx ...*gorm.DB) ([]*model.Domain, error) {
	list := make([]*model.Domain, 0)
	if len(siteIds) == 0 {
		return list, nil
	}
	err := r.db(tx...).Where("site_id IN ?", siteIds).Order("site_id ASC, id ASC").Find(&list).Error
	return list, err
}

// CountByCertId 查询证书被域名引用的数量。
func (r *Repository) CountByCertId(certId int64, tx ...*gorm.DB) (int64, error) {
	var count int64
	err := r.db(tx...).Model(&model.Domain{}).Where("cert_id = ?", certId).Count(&count).Error
	return count, err
}
