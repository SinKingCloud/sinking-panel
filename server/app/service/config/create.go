package config

import (
	"server/app/model"
	"server/app/util"
	"server/app/util/str"
	"time"
)

// Create 创建数据
func (s *Service) Create(key string, value string) (err error) {
	err = util.Database.Db.Create(&model.Config{
		Key:        key,
		Value:      value,
		CreateTime: str.DateTime(time.Now()),
	}).Error
	return
}
