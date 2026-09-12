package log

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 获取数据
func (r *Repository) Select(where *SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error) {
	query := r.Database.Db.Model(&model.Log{})
	if where != nil {
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("(ip LIKE ? OR location LIKE ? OR title LIKE ? OR content LIKE ?)", keyword, keyword, keyword, keyword)
		}
		if where.Type != nil {
			query = query.Where("type = ?", *where.Type)
		}
		if where.Ip != nil {
			query = query.Where("ip LIKE ?", "%"+*where.Ip+"%")
		}
		if where.Location != nil {
			query = query.Where("location LIKE ?", "%"+*where.Location+"%")
		}
		if where.Title != nil {
			query = query.Where("title LIKE ?", "%"+*where.Title+"%")
		}
		if where.Content != nil {
			query = query.Where("content LIKE ?", "%"+*where.Content+"%")
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
