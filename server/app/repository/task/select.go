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
		if where.TypeId != nil {
			query = query.Where("type_id = ?", *where.TypeId)
		}
		if where.ExecType != nil {
			query = query.Where("exec_type = ?", *where.ExecType)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Status != nil {
			query = query.Where("status = ?", *where.Status)
		}
		if where.RunTimeStart != nil {
			query = query.Where("run_time >= ?", *where.RunTimeStart)
		}
		if where.RunTimeEnd != nil {
			query = query.Where("run_time <= ?", *where.RunTimeEnd)
		}
		if where.CreateTimeStart != nil {
			query = query.Where("create_time >= ?", *where.CreateTimeStart)
		}
		if where.CreateTimeEnd != nil {
			query = query.Where("create_time <= ?", *where.CreateTimeEnd)
		}
		if where.UpdateTimeStart != nil {
			query = query.Where("update_time >= ?", *where.UpdateTimeStart)
		}
		if where.UpdateTimeEnd != nil {
			query = query.Where("update_time <= ?", *where.UpdateTimeEnd)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
