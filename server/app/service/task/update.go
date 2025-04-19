package task

import (
	"server/app/model"
	"server/app/util"
	"server/app/util/str"
	"time"
)

// updateEntryIDById 通过ID更新entryID
func (s *Service) updateEntryIDById(id int, entryID int) (err error) {
	data := map[string]interface{}{
		"update_time": str.DateTime(time.Now()),
		"entry_id":    entryID,
	}
	err = util.Database.Db.Model(&model.Task{}).Where("`id` = ? ", id).Updates(data).Error
	return
}

// updateStatusById 通过ID更新状态
func (s *Service) updateStatusById(id int, status Status) (err error) {
	data := map[string]interface{}{
		"update_time": str.DateTime(time.Now()),
		"status":      int(status),
	}
	err = util.Database.Db.Model(&model.Task{}).Where("`id` = ? ", id).Updates(data).Error
	return
}

// updateStatusById 通过ID更新运行时间
func (s *Service) updateRuntimeById(id int, t time.Time) (err error) {
	data := map[string]interface{}{
		"update_time": str.DateTime(time.Now()),
		"run_time":    str.DateTime(t),
	}
	err = util.Database.Db.Model(&model.Task{}).Where("`id` = ? ", id).Updates(data).Error
	return
}

// UpdateByIds 通过ID列表更新
func (s *Service) UpdateByIds(ids []int, data map[string]interface{}) (err error) {
	data["update_time"] = str.DateTime(time.Now())
	err = util.Database.Db.Model(&model.Task{}).Where("`id` in ? ", ids).Updates(data).Error
	if err == nil {
		for _, v := range ids {
			_ = s.Refresh(v)
		}
	}
	return
}
