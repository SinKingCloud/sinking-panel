package site_domain

import (
	"errors"
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateCertId 修改域名使用的证书，0 表示关闭 SSL。
func (r *Repository) UpdateCertId(id, certId int64, tx ...*gorm.DB) error {
	if certId < 0 {
		return errors.New("证书 ID 不能小于 0")
	}
	execute := func(db *gorm.DB) error {
		result := db.Model(&model.SiteDomain{}).Where("id = ?", id).Updates(map[string]interface{}{
			"cert_id":     certId,
			"update_time": str.DateTime(time.Now()),
		})
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
