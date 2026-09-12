package site

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateById 更新网站。
func (r *Repository) UpdateById(id int64, data *UpdateSite, tx ...*gorm.DB) error {
	if data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.TypeId != nil {
		updates["type_id"] = *data.TypeId
	}
	if data.Type != nil {
		updates["type"] = *data.Type
	}
	if data.Status != nil {
		updates["status"] = *data.Status
	}
	if data.Root != nil {
		updates["root"] = *data.Root
	}
	if data.RunPath != nil {
		updates["run_path"] = *data.RunPath
	}
	if data.Config != nil {
		updates["config"] = *data.Config
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	execute := func(db *gorm.DB) error {
		result := db.Model(&model.Site{}).
			Where("id = ?", id).
			Updates(updates)
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

// ClearTypeId 清除指定类型的网站分类。
func (r *Repository) ClearTypeId(typeIds []int64, tx ...*gorm.DB) error {
	if len(typeIds) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(typeIds, 1000, func(batch interface{}) error {
			return db.Model(&model.Site{}).
				Where("type_id IN ?", batch.([]int64)).
				Updates(map[string]interface{}{
					"type_id":     int64(0),
					"update_time": str.DateTime(time.Now()),
				}).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return execute(tx)
	})
}
