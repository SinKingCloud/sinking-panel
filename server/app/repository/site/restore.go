package site

import (
	"errors"
	"fmt"
	"server/app/model"

	"gorm.io/gorm"
)

// RestoreRecords 重新写入网站及域名快照，保留原始时间，供运行时同步失败后补偿。
func (r *Repository) RestoreRecords(data *model.Site, domains []*model.SiteDomain, tx ...*gorm.DB) error {
	if data == nil || data.Id <= 0 {
		return errors.New("缺少待恢复的网站数据")
	}
	website := *data
	list := make([]*model.SiteDomain, 0, len(domains))
	for _, item := range domains {
		if item != nil {
			domain := *item
			list = append(list, &domain)
		}
	}
	execute := func(db *gorm.DB) error {
		if err := db.Where("site_id = ?", data.Id).Delete(&model.SiteDomain{}).Error; err != nil {
			return fmt.Errorf("补偿清理网站域名失败: %w", err)
		}
		if err := db.Where("id = ?", data.Id).Delete(&model.Site{}).Error; err != nil {
			return fmt.Errorf("补偿清理网站失败: %w", err)
		}
		withoutHooks := db.Session(&gorm.Session{SkipHooks: true})
		if err := withoutHooks.Create(&website).Error; err != nil {
			return fmt.Errorf("补偿恢复网站失败: %w", err)
		}
		if err := r.Database.BatchExecute(list, 500, func(batch interface{}) error {
			return withoutHooks.Create(batch.([]*model.SiteDomain)).Error
		}); err != nil {
			return fmt.Errorf("补偿恢复网站域名失败: %w", err)
		}
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return execute(tx[0])
	}
	return r.Database.Transaction(execute)
}
