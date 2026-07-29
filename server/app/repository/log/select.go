package log

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 获取数据
func (r *Repository) Select(where *SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error) {
	query := r.Database.Db.Model(&model.Log{})
	if where != nil {
		if where.Type != "" {
			query = query.Where("type = ?", where.Type)
		}
		if where.Ip != "" {
			query = query.Where("ip LIKE ?", "%"+where.Ip+"%")
		}
		if where.Title != "" {
			query = query.Where("title LIKE ?", "%"+where.Title+"%")
		}
		if where.Content != "" {
			query = query.Where("content LIKE ?", "%"+where.Content+"%")
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
