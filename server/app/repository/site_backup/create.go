package site_backup

import "server/app/model"

// Create 保存已生成的网站备份。
func (r *Repository) Create(data *model.SiteBackup) error {
	return r.Database.Db.Create(data).Error
}
