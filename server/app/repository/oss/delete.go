package oss

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteByIds 批量删除面板连接记录，不操作云端存储桶或对象。
func (r *Repository) DeleteByIds(ids []int64, tx ...*gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return db.Where("id IN ?", batch.([]int64)).Delete(&model.Oss{}).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(execute)
}
