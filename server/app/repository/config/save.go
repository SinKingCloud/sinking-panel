package config

import (
	"server/app/model"
	"server/app/util/str"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Save 保存数据
func (r *Repository) Save(configs []*model.Config) error {
	if len(configs) == 0 {
		return nil
	}
	return r.Database.Transaction(func(tx *gorm.DB) error {
		return r.Database.BatchExecute(configs, 1000, func(batch interface{}) error {
			batchConfigs := batch.([]*model.Config)
			return tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "key"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"value":       gorm.Expr("excluded.value"),
					"update_time": str.DateTime(time.Now()),
				}),
			}).CreateInBatches(batchConfigs, len(batchConfigs)).Error
		})
	})
}
