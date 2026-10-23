package database

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateById 更新数据库实例。
func (r *Repository) UpdateById(id int64, data *UpdateDatabase, tx ...*gorm.DB) error {
	if data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.TypeId != nil {
		updates["type_id"] = *data.TypeId
	}
	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.Config != nil {
		updates["config"] = *data.Config
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	result := db.Model(&model.Database{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ClearTypeId 清除被删除的实例分类。
func (r *Repository) ClearTypeId(typeIds []int64, tx ...*gorm.DB) error {
	if len(typeIds) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(typeIds, 1000, func(batch interface{}) error {
			return db.Model(&model.Database{}).
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
	return r.Database.Transaction(execute)
}
