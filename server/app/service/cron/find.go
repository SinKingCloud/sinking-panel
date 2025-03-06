package cron

import (
	"server/app/model"
	"server/app/util"
)

// findById 通过ID查询
func (s *Service) findById(id int) (data *model.Task, err error) {
	err = util.Database.Db.Where("`id` = ? ", id).First(&data).Error
	return
}

// FindById 通过ID查询
func (s *Service) FindById(id int) (data *model.Task, err error) {
	return s.findById(id)
}
