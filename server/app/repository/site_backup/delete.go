package site_backup

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteByIds 批量删除备份记录，文件清理由业务层处理。
func (r *Repository) DeleteByIds(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return tx.Where("id IN ?", batch.([]int64)).Delete(&model.SiteBackup{}).Error
		})
	})
}
