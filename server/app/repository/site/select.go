package site

import (
	"server/app/model"
	"server/app/util/page"

	"gorm.io/gorm"
)

// SelectAll 查询全部网站及其配置。
func (r *Repository) SelectAll(tx ...*gorm.DB) ([]*model.Site, error) {
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	list := make([]*model.Site, 0)
	err := db.Order("id ASC").Find(&list).Error
	return list, err
}

// Select 分页查询网站。
func (r *Repository) Select(where *SelectSite, queryPage *page.Query) (*page.Result[*Site], error) {
	query := r.Database.Db.Model(&model.Site{})
	if where != nil {
		if where.TypeId != nil {
			query = query.Where("type_id = ?", *where.TypeId)
		}
		if where.Type != nil {
			query = query.Where("type = ?", *where.Type)
		}
		if where.Status != nil {
			query = query.Where("status = ?", *where.Status)
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
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("name LIKE ? OR root LIKE ? OR run_path LIKE ?", keyword, keyword, keyword)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Root != nil {
			query = query.Where("root LIKE ?", "%"+*where.Root+"%")
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
