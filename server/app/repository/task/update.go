package task

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateEntryIDById 更新任务实例ID
func (r *Repository) UpdateEntryIDById(id int64, entryID int) error {
	return r.Database.Db.Model(&model.Task{}).Where("`id` = ? ", id).Updates(map[string]interface{}{
		"update_time": str.DateTime(time.Now()),
		"entry_id":    entryID,
	}).Error
}

// UpdateStateById 更新任务实例和状态
func (r *Repository) UpdateStateById(id int64, entryID int, status int) error {
	return r.Database.Db.Model(&model.Task{}).Where("`id` = ? ", id).Updates(map[string]interface{}{
		"update_time": str.DateTime(time.Now()),
		"entry_id":    entryID,
		"status":      status,
	}).Error
}

// UpdateRuntimeById 更新任务运行时间
func (r *Repository) UpdateRuntimeById(id int64, runTime time.Time) error {
	return r.Database.Db.Model(&model.Task{}).Where("`id` = ? ", id).Updates(map[string]interface{}{
		"update_time": str.DateTime(time.Now()),
		"run_time":    str.DateTime(runTime),
	}).Error
}

// UpdateByIds 通过ID列表更新任务
func (r *Repository) UpdateByIds(ids []int64, data *UpdateTask) error {
	if len(ids) == 0 || data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Type != nil {
		updates["type"] = data.Type
	}
	if data.Name != nil {
		updates["name"] = data.Name
	}
	if data.Spec != nil {
		updates["spec"] = data.Spec
	}
	if data.Script != nil {
		updates["script"] = data.Script
	}
	if data.Status != nil {
		updates["status"] = data.Status
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return tx.Model(&model.Task{}).
				Where("id IN ?", batch.([]int64)).
				Updates(updates).Error
		})
	})
}
