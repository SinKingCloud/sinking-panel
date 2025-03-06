package server

import (
	"server/app/model"
	"server/app/util"
)

// Create 插入数据
func (s *Service) Create(data *model.Server) (err error) {
	err = util.Database.Db.Create(&data).Error
	return
}
