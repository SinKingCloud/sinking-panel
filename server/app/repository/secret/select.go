package secret

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 分页查询密钥凭据。
func (r *Repository) Select(where *SelectSecret, queryPage *page.Query) (*page.Result[*Secret], error) {
	query := r.Database.Db.Model(&model.Secret{})
	if where != nil {
		if where.Keyword != nil {
			query = query.Where("name LIKE ?", "%"+*where.Keyword+"%")
		}
		if where.Provider != nil {
			query = query.Where("provider = ?", *where.Provider)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}

// SelectIdNameMap 查询密钥 ID 和名称，不读取凭据正文。
func (r *Repository) SelectIdNameMap(where *SelectSecret) (map[int64]string, error) {
	query := r.Database.Db.Model(&model.Secret{})
	if where != nil {
		if where.Keyword != nil {
			query = query.Where("name LIKE ?", "%"+*where.Keyword+"%")
		}
		if where.Provider != nil {
			query = query.Where("provider = ?", *where.Provider)
		}
	}
	var data []struct {
		Id   int64
		Name string
	}
	err := query.
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
