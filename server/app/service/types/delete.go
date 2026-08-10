package types

import (
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// DeleteByIds 通过ID列表删除
func (s *service) DeleteByIds(ids []int64) error {
	return s.database.Transaction(func(tx *gorm.DB) error {
		return s.database.BatchExecute(ids, 1000, func(batch interface{}) error {
			batchIds := batch.([]int64)
			typeIds := tx.Model(&model.Type{}).
				Select("id").
				Where("id IN ? AND module = ?", batchIds, type_module.Script)
			if err := tx.Model(&model.Script{}).
				Where("type_id IN (?)", typeIds).
				Updates(map[string]interface{}{
					"type_id":     int64(0),
					"update_time": str.DateTime(time.Now()),
				}).Error; err != nil {
				return err
			}
			return s.repositoryTypes.DeleteByIds(batchIds, tx)
		})
	})
}
