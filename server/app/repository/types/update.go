package types

import (
	"server/app/model"
	"server/app/util/str"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UpdateByIds 通过ID列表更新
func (r *Repository) UpdateByIds(ids []int64, data *UpdateType, tx ...*gorm.DB) error {
	if len(ids) == 0 || data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Module != nil {
		updates["module"] = *data.Module
	}
	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.Sort != nil {
		updates["sort"] = *data.Sort
	}
	reorder := data.Reorder != nil && *data.Reorder
	if len(updates) == 0 && !reorder {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	execute := func(db *gorm.DB) error {
		position := int64(1)
		return r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
			batchIds := batch.([]int64)
			if reorder {
				// 每批一次更新，所有批次沿用同一事务和连续位置，避免相邻 sort 值碰撞。
				var expression strings.Builder
				expression.WriteString("CASE id")
				values := make([]interface{}, 0, len(batchIds)*2)
				for _, id := range batchIds {
					expression.WriteString(" WHEN ? THEN ?")
					values = append(values, id, position)
					position++
				}
				expression.WriteString(" END")
				updates["sort"] = gorm.Expr(expression.String(), values...)
			}
			return db.Model(&model.Type{}).
				Where("id IN ?", batchIds).
				Updates(updates).Error
		})
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return execute(tx)
	})
}
