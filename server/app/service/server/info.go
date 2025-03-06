package server

import (
	"server/app/model"
	"server/app/util"
)

// FindById 查询信息
func (*Service) FindById(id int) (user *model.Server, err error) {
	err = util.Database.Db.Where("`id` = ?", id).First(&user).Error
	return
}
