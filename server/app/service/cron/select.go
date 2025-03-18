package cron

import (
	"server/app/model"
	"server/app/util"
)

// SelectAll 获取所有数据
func (s *Service) selectAll() (list []*model.Task, err error) {
	err = util.Database.Db.Find(&list).Error
	return
}

// Select 获取数据
func (s *Service) Select(where map[string]string, orderByField string, orderByType string, page int, pageSize int) (list []*Task, total int64, err error) {
	offset := pageSize * (page - 1)
	query := util.Database.Db.Model(&model.Task{})
	if where["type"] != "" {
		query.Where("`type` = ?", where["type"])
	}
	if where["name"] != "" {
		query.Where("`name` like ?", "%"+where["name"]+"%")
	}
	if where["status"] != "" {
		query.Where("`status` = ?", where["status"])
	}
	if where["run_time_start"] != "" {
		query.Where("`run_time` >= ?", where["run_time_start"])
	}
	if where["run_time_end"] != "" {
		query.Where("`run_time` <= ?", where["run_time_end"])
	}
	if where["create_time_start"] != "" {
		query.Where("`create_time` >= ?", where["create_time_start"])
	}
	if where["create_time_end"] != "" {
		query.Where("`create_time` <= ?", where["create_time_end"])
	}
	if where["update_time_start"] != "" {
		query.Where("`update_time` >= ?", where["update_time_start"])
	}
	if where["update_time_end"] != "" {
		query.Where("`update_time` <= ?", where["update_time_end"])
	}
	err = query.Count(&total).Limit(pageSize).Offset(offset).Order(orderByField + " " + orderByType).Find(&list).Error
	return
}
