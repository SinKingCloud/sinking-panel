package software

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 分页查询软件，不读取配置及关联资源。
func (r *Repository) Select(where *SelectSoftware, queryPage *page.Query) (*page.Result[*Software], error) {
	query := r.Database.Db.Model(&model.Software{}).
		Select("id", "name", "code", "version", "status", "state", "create_time", "update_time")
	if where != nil {
		if where.Code != nil {
			query = query.Where("code = ?", *where.Code)
		}
		if where.Status != nil {
			query = query.Where("status = ?", *where.Status)
		}
		if where.State != nil {
			query = query.Where("state = ?", *where.State)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("(name LIKE ? OR code LIKE ?)", keyword, keyword)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}

// SelectAll 查询启动恢复所需的软件。
func (r *Repository) SelectAll() ([]*model.Software, error) {
	data := make([]*model.Software, 0)
	err := r.Database.Db.Omit("config").Order("id asc").Find(&data).Error
	return data, err
}
