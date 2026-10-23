package database

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 分页查询数据库实例。
func (r *Repository) Select(where *SelectDatabase, queryPage *page.Query) (*page.Result[*Database], error) {
	query := r.Database.Db.Model(&model.Database{}).
		Select("id", "type_id", "name", "type", "create_time", "update_time", "CASE WHEN config <> '' THEN 1 ELSE 0 END AS configured")
	if where != nil {
		if where.Type != nil {
			query = query.Where("type = ?", *where.Type)
		}
		if where.TypeId != nil {
			query = query.Where("type_id = ?", *where.TypeId)
		}
		if where.Keyword != nil {
			query = query.Where("name LIKE ?", "%"+*where.Keyword+"%")
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
