package secret

import (
	"server/app/model"
	"server/app/util/page"

	"gorm.io/gorm"
)

// Select 分页查询密钥凭据。
func (r *Repository) Select(where *SelectSecret, queryPage *page.Query) (*page.Result[*Secret], error) {
	return r.Repository.SelectPage(r.query(where), queryPage)
}

// SelectIdNameMap 查询密钥 ID 和名称，不读取凭据正文。
func (r *Repository) SelectIdNameMap(where *SelectSecret) (map[int64]string, error) {
	var data []struct {
		Id   int64
		Name string
	}
	err := r.query(where).
		Select("id", "name").
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

func (r *Repository) query(where *SelectSecret) *gorm.DB {
	query := r.Database.Db.Model(&model.Secret{})
	if where != nil {
		if where.Keyword != "" {
			query = query.Where("name LIKE ?", "%"+where.Keyword+"%")
		}
		if where.Provider != "" {
			query = query.Where("provider = ?", where.Provider)
		}
	}
	return query
}
