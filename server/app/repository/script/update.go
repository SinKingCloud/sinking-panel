package script

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateByIds 通过ID列表更新常用脚本
func (r *Repository) UpdateByIds(ids []int64, data *UpdateScript) error {
	if len(ids) == 0 || data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.TypeId != nil {
		updates["type_id"] = data.TypeId
	}
	if data.Name != nil {
		updates["name"] = data.Name
	}
	if data.Script != nil {
		updates["script"] = data.Script
	}
	if data.Sort != nil {
		updates["sort"] = data.Sort
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return tx.Model(&model.Script{}).
				Where("id IN ?", batch.([]int64)).
				Updates(updates).Error
		})
	})
}
