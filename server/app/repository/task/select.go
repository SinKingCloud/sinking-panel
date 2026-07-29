package task

import (
	"server/app/model"
	"server/app/util/page"
)

// SelectAll 获取所有任务
func (r *Repository) SelectAll() (list []*model.Task, err error) {
	err = r.Database.Db.Find(&list).Error
	return
}

// Select 查询任务
func (r *Repository) Select(where *SelectTask, queryPage *page.Query) (*page.Result[*Task], error) {
	query := r.Database.Db.Model(&model.Task{})
	if where != nil {
		if where.Type != "" {
			query = query.Where("type = ?", where.Type)
		}
		if where.Name != "" {
			query = query.Where("name LIKE ?", "%"+where.Name+"%")
		}
		if where.Status != "" {
			query = query.Where("status = ?", where.Status)
		}
		if where.RunTimeStart != "" {
			query = query.Where("run_time >= ?", where.RunTimeStart)
		}
		if where.RunTimeEnd != "" {
			query = query.Where("run_time <= ?", where.RunTimeEnd)
		}
		if where.CreateTimeStart != "" {
			query = query.Where("create_time >= ?", where.CreateTimeStart)
		}
		if where.CreateTimeEnd != "" {
			query = query.Where("create_time <= ?", where.CreateTimeEnd)
		}
		if where.UpdateTimeStart != "" {
			query = query.Where("update_time >= ?", where.UpdateTimeStart)
		}
		if where.UpdateTimeEnd != "" {
			query = query.Where("update_time <= ?", where.UpdateTimeEnd)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
