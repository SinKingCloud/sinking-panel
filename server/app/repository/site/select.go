package site

import (
	"server/app/enum/site_status"
	"server/app/model"
	"server/app/util/page"

	"gorm.io/gorm"
)

// SelectIdNameMap 查询可选默认站点的 ID 名称映射。
func (r *Repository) SelectIdNameMap() (map[int64]string, error) {
	var data []struct {
		Id   int64
		Name string
	}
	err := r.Database.Db.Model(&model.Site{}).
		Select("id", "name").
		Where("status = ?", site_status.Enabled).
		Where("id IN (?)", r.Database.Db.Model(&model.SiteDomain{}).Select("site_id")).
		Order("id ASC").
		Find(&data).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]string, len(data))
	for _, item := range data {
		result[item.Id] = item.Name
	}
	return result, nil
}

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
		if where.Keyword != "" {
			keyword := "%" + where.Keyword + "%"
			query = query.Where("name LIKE ? OR root LIKE ?", keyword, keyword)
		}
		if where.Name != "" {
			query = query.Where("name LIKE ?", "%"+where.Name+"%")
		}
		if where.TypeId != "" {
			query = query.Where("type_id = ?", where.TypeId)
		}
		if where.Type != "" {
			query = query.Where("type = ?", where.Type)
		}
		if where.Status != "" {
			query = query.Where("status = ?", where.Status)
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
