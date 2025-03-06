package cron

import (
	"server/app/model"
	"server/app/util"
)

// create 插入数据
func (s *Service) create(data *model.Task) (err error) {
	err = util.Database.Db.Create(&data).Error
	return
}
