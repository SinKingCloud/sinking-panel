package cert

import (
	"server/app/model"
	"server/app/util/page"

	"gorm.io/gorm"
)

// SelectByIds 查询指定证书。
func (r *Repository) SelectByIds(ids []int64, tx ...*gorm.DB) ([]*model.Cert, error) {
	list := make([]*model.Cert, 0)
	if len(ids) == 0 {
		return list, nil
	}
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("id IN ?", ids).Order("id ASC").Find(&list).Error
	return list, err
}

// Select 分页查询证书。
func (r *Repository) Select(where *SelectCert, queryPage *page.Query) (*page.Result[*Cert], error) {
	query := r.Database.Db.Model(&model.Cert{})
	if where != nil {
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("name LIKE ? OR domains LIKE ?", keyword, keyword)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Type != nil {
			query = query.Where("type = ?", *where.Type)
		}
		if where.ExpireTimeStart != nil {
			query = query.Where("expire_time >= ?", *where.ExpireTimeStart)
		}
		if where.ExpireTimeEnd != nil {
			query = query.Where("expire_time <= ?", *where.ExpireTimeEnd)
		}
		if where.CreateTimeStart != nil {
			query = query.Where("create_time >= ?", *where.CreateTimeStart)
		}
		if where.CreateTimeEnd != nil {
			query = query.Where("create_time <= ?", *where.CreateTimeEnd)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
