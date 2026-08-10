package types

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateByIds 通过ID列表更新
func (r *Repository) UpdateByIds(ids []int64, data *UpdateType, tx ...*gorm.DB) error {
	if len(ids) == 0 || data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Module != nil {
		updates["module"] = data.Module
	}
	if data.Name != nil {
		updates["name"] = data.Name
	}
	if data.Sort != nil {
		updates["sort"] = data.Sort
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return db.Model(&model.Type{}).
				Where("id IN ?", batch.([]int64)).
				Updates(updates).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return execute(tx)
	})
}
