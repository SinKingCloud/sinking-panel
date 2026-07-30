package task

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteByIds 通过ID列表删除任务
func (r *Repository) DeleteByIds(ids []int64, tx ...*gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return db.Where("id IN ?", batch.([]int64)).Delete(&model.Task{}).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return execute(tx)
	})
}
