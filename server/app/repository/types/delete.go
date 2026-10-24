package types

import (
	"server/app/model"

	"gorm.io/gorm"
)

// DeleteByIds 通过ID列表删除
func (r *Repository) DeleteByIds(ids []int64, tx ...*gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			return db.Where("id IN ?", batch.([]int64)).Delete(&model.Type{}).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return execute(tx)
	})
}

// DeleteByModules 分批删除指定模块的分类。
func (r *Repository) DeleteByModules(modules []string, tx ...*gorm.DB) error {
	if len(modules) == 0 {
		return nil
	}
	execute := func(db *gorm.DB) error {
		return r.Database.BatchExecute(modules, 1000, func(batch interface{}) error {
			return db.Where("module IN ?", batch.([]string)).Delete(&model.Type{}).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(execute)
}
