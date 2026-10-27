package types

import (
	"server/app/model"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SelectByIds 查询指定ID的类型
func (r *Repository) SelectByIds(ids []int64, tx ...*gorm.DB) ([]*model.Type, error) {
	result := make([]*model.Type, 0)
	if len(ids) == 0 {
		return result, nil
	}
	db := r.Database.Db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
		var data []*model.Type
		if err := db.Model(&model.Type{}).
			Select("id", "module").
			Where("id IN ?", batch.([]int64)).
			Find(&data).Error; err != nil {
			return err
		}
		result = append(result, data...)
		return nil
	})
	return result, err
}

// SelectIdNameMap 查询指定模块的ID名称映射
func (r *Repository) SelectIdNameMap(module string) (map[int64]string, error) {
	var data []struct {
		Id   int64  `gorm:"column:id" json:"id"`
		Name string `gorm:"column:name" json:"name"`
	}
	query := r.Database.Db.Model(&model.Type{}).Select("id", "name")
	if strings.Contains(module, "%") {
		query = query.Where("module LIKE ? ESCAPE '!'", strings.NewReplacer("!", "!!", "_", "!_").Replace(module))
	} else {
		query = query.Where("module = ?", module)
	}
	err := query.Order("sort ASC, id ASC").Find(&data).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]string, len(data))
	for _, item := range data {
		result[item.Id] = item.Name
	}
	return result, nil
}

// Select 获取数据
func (r *Repository) Select(where *SelectType, orderByField, orderByType string) ([]*model.Type, error) {
	query := r.Database.Db.Model(&model.Type{})
	if where != nil {
		if where.Module != nil {
			if strings.Contains(*where.Module, "%") {
				query = query.Where("module LIKE ? ESCAPE '!'", strings.NewReplacer("!", "!!", "_", "!_").Replace(*where.Module))
			} else if module, _, scoped := strings.Cut(*where.Module, "_"); scoped {
				query = query.Where("module IN ?", []string{module, *where.Module})
			} else {
				query = query.Where("module = ?", *where.Module)
			}
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
	}
	if orderByField != "id" {
		orderByField = "sort"
	}
	desc := orderByType == "desc"
	order := []clause.OrderByColumn{{
		Column: clause.Column{Name: orderByField},
		Desc:   desc,
	}}
	if orderByField != "id" {
		order = append(order, clause.OrderByColumn{
			Column: clause.Column{Name: "id"},
			Desc:   desc,
		})
	}
	data := make([]*model.Type, 0)
	err := query.Clauses(clause.OrderBy{Columns: order}).Find(&data).Error
	return data, err
}
