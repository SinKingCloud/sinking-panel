package site_backup

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 分页查询网站备份，网站删除后仍保留备份记录。
func (r *Repository) Select(where *SelectSiteBackup, queryPage *page.Query) (*page.Result[*SiteBackup], error) {
	backup := &model.SiteBackup{}
	site := &model.Site{}
	query := r.Database.Db.Model(backup).Table(backup.TableName() + " AS backup").
		Select("backup.*, site.name AS site_name").
		Joins("LEFT JOIN " + site.TableName() + " AS site ON site.id = backup.site_id")
	if where != nil {
		if where.SiteId != nil {
			query = query.Where("backup.site_id = ?", *where.SiteId)
		}
		if where.CreateTimeStart != nil {
			query = query.Where("backup.create_time >= ?", *where.CreateTimeStart)
		}
		if where.CreateTimeEnd != nil {
			query = query.Where("backup.create_time <= ?", *where.CreateTimeEnd)
		}
		if where.Name != nil {
			query = query.Where("backup.name LIKE ?", "%"+*where.Name+"%")
		}
		if where.SiteName != nil {
			query = query.Where("site.name LIKE ?", "%"+*where.SiteName+"%")
		}
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("(backup.name LIKE ? OR site.name LIKE ?)", keyword, keyword)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}

// SelectBySiteId 查询网站的全部备份，按 ID 倒序排列。
func (r *Repository) SelectBySiteId(id int64) ([]*model.SiteBackup, error) {
	list := make([]*model.SiteBackup, 0)
	err := r.Database.Db.Where("site_id = ?", id).Order("id DESC").Find(&list).Error
	return list, err
}

// SelectByIds 批量查询备份记录。
func (r *Repository) SelectByIds(ids []int64) ([]*model.SiteBackup, error) {
	list := make([]*model.SiteBackup, 0)
	err := r.Database.BatchExecute(ids, 1000, func(batch interface{}) error {
		var data []*model.SiteBackup
		if err := r.Database.Db.Where("id IN ?", batch.([]int64)).Find(&data).Error; err != nil {
			return err
		}
		list = append(list, data...)
		return nil
	})
	return list, err
}
