package secret

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 分页查询密钥凭据。
func (r *Repository) Select(where *SelectSecret, queryPage *page.Query) (*page.Result[*Secret], error) {
	query := r.Database.Db.Model(&model.Secret{})
	if where != nil {
		if where.Provider != nil {
			query = query.Where("provider = ?", *where.Provider)
		}
		if where.Keyword != nil {
			query = query.Where("name LIKE ?", "%"+*where.Keyword+"%")
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}

// SelectIdNameMapByIds 分批读取指定密钥的名称，不加载凭据内容。
func (r *Repository) SelectIdNameMapByIds(ids []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	err := r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
		var records []*model.Secret
		if err := r.Database.Db.Model(&model.Secret{}).
			Select("id", "name").
			Where("id IN ?", batch.([]int64)).
			Find(&records).Error; err != nil {
			return err
		}
		for _, record := range records {
			result[record.Id] = record.Name
		}
		return nil
	})
	return result, err
}
