package cron

import (
	"server/app/model"
	"server/app/util"
)

// deleteById 通过ID删除
func (s *Service) deleteById(id int) (err error) {
	err = util.Database.Db.Where("`id` = ? ", id).Delete(&model.Task{}).Error
	return
}
