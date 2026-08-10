package types

import (
	"errors"
	"server/app/enum/type_module"
	"server/app/model"
	repositoryTypes "server/app/repository/types"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
)

// UpdateByIds 通过ID列表更新
func (s *service) UpdateByIds(ids []int64, data *repositoryTypes.UpdateType) error {
	if len(ids) == 0 || data == nil {
		return errors.New("更新数据不能为空")
	}
	if data.Module != nil {
		value, ok := data.Module.(string)
		if !ok {
			return errors.New("所属模块不合法")
		}
		if _, ok = type_module.Map()[value]; !ok {
			return errors.New("所属模块不合法")
		}
		if value != type_module.Script {
			return s.database.Transaction(func(tx *gorm.DB) error {
				return s.database.BatchExecute(ids, 1000, func(batch interface{}) error {
					batchIds := batch.([]int64)
					typeIds := tx.Model(&model.Type{}).
						Select("id").
						Where("id IN ? AND module = ?", batchIds, type_module.Script)
					if err := tx.Model(&model.Script{}).
						Where("type_id IN (?)", typeIds).
						Updates(map[string]interface{}{
							"type_id":     int64(0),
							"update_time": str.DateTime(time.Now()),
						}).Error; err != nil {
						return err
					}
					return s.repositoryTypes.UpdateByIds(batchIds, data, tx)
				})
			})
		}
	}
	return s.repositoryTypes.UpdateByIds(ids, data)
}
