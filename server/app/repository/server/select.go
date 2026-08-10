package server

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 获取数据
func (r *Repository) Select(where *SelectServer, queryPage *page.Query) (*page.Result[*Server], error) {
	query := r.Database.Db.Model(&model.Server{})
	if where != nil {
		if where.Keyword != "" {
			keyword := "%" + where.Keyword + "%"
			query = query.Where("(name LIKE ? OR ip LIKE ? OR CAST(port AS TEXT) LIKE ?)", keyword, keyword, keyword)
		}
		if where.Ip != "" {
			query = query.Where("ip LIKE ?", "%"+where.Ip+"%")
		}
		if where.Port != "" {
			query = query.Where("port = ?", where.Port)
		}
		if where.User != "" {
			query = query.Where("user LIKE ?", "%"+where.User+"%")
		}
		if where.AuthType != "" {
			query = query.Where("auth_type = ?", where.AuthType)
		}
		if where.Name != "" {
			query = query.Where("name LIKE ?", "%"+where.Name+"%")
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
