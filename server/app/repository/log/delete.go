package log

import (
	"time"

	"server/app/model"

	"gorm.io/gorm"
)

// SelectIdByCreateTime 查询指定时间之前的日志ID。
func (r *Repository) SelectIdByCreateTime(before time.Time, limit int) ([]int64, error) {
	if limit <= 0 {
		return []int64{}, nil
	}
	ids := make([]int64, 0, limit)
	err := r.Database.Db.Model(&model.Log{}).
		Select("id").
		Where("create_time < ?", before).
		Order("id asc").
		Limit(limit).
		Pluck("id", &ids).Error
	return ids, err
}

// Delete 按ID批量删除操作日志。
func (r *Repository) Delete(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return tx.Where("id IN ?", batch.([]int64)).Delete(&model.Log{}).Error
		})
	})
}
