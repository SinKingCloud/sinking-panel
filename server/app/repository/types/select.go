package types

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 获取数据
func (r *Repository) Select(where *SelectType, queryPage *page.Query) (*page.Result[*model.Type], error) {
	query := r.Database.Db.Model(&model.Type{})
	if where != nil {
		if where.Module != "" {
			query = query.Where("module = ?", where.Module)
		}
		if where.Name != "" {
			query = query.Where("name LIKE ?", "%"+where.Name+"%")
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
