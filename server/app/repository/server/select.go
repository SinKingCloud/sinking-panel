package server

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 获取数据
func (r *Repository) Select(where *SelectServer, queryPage *page.Query) (*page.Result[*Server], error) {
	query := r.Database.Db.Model(&model.Server{})
	if where != nil {
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("(name LIKE ? OR ip LIKE ? OR CAST(port AS TEXT) LIKE ?)", keyword, keyword, keyword)
		}
		if where.Ip != nil {
			query = query.Where("ip LIKE ?", "%"+*where.Ip+"%")
		}
		if where.Port != nil {
			query = query.Where("port = ?", *where.Port)
		}
		if where.User != nil {
			query = query.Where("user LIKE ?", "%"+*where.User+"%")
		}
		if where.AuthType != nil {
			query = query.Where("auth_type = ?", *where.AuthType)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
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
