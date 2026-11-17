package site_backup

import "server/app/model"

// UpdateData 按原值更新副本信息，返回实际更新数，避免并行清理覆盖已删除的副本。
func (r *Repository) UpdateData(id int64, before, data string) (int64, error) {
	result := r.Database.Db.Model(&model.SiteBackup{}).Where("id = ? AND data = ?", id, before).Update("data", data)
	return result.RowsAffected, result.Error
}
