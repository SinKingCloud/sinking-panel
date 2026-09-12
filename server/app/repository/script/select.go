package script

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 查询常用脚本
func (r *Repository) Select(where *SelectScript, queryPage *page.Query) (*page.Result[*Script], error) {
	query := r.Database.Db.Model(&model.Script{})
	if where != nil {
		if where.TypeId != nil {
			query = query.Where("type_id = ?", *where.TypeId)
		}
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("(name LIKE ? OR script LIKE ?)", keyword, keyword)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Script != nil {
			query = query.Where("script LIKE ?", "%"+*where.Script+"%")
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
