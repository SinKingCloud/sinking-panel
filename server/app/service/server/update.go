package server

import (
	"server/app/model"
	"server/app/util"
	"server/app/util/str"
	"time"
)

// UpdateByIds 通过ID列表更新
func (s *Service) UpdateByIds(ids []int, data map[string]interface{}) (err error) {
	data["update_time"] = str.DateTime(time.Now())
	err = util.Database.Db.Model(&model.Server{}).Where("`id` in ? ", ids).Updates(data).Error
	return
}
