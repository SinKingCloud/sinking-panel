package server

import (
	"server/app/model"
	"server/app/util"
)

// DeleteByIds 通过ID列表删除
func (s *Service) DeleteByIds(ids []int) (err error) {
	err = util.Database.Db.Where("`id` in ? ", ids).Delete(&model.Server{}).Error
	return
}
