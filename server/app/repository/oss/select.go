package oss

import (
	"server/app/model"
	"server/app/util/page"
)

// Select 分页查询对象存储基本信息，不读取连接配置。
func (r *Repository) Select(where *SelectOss, queryPage *page.Query) (*page.Result[*Oss], error) {
	query := r.Database.Db.Model(&model.Oss{})
	if where != nil {
		if where.Keyword != nil {
			query = query.Where("name LIKE ?", "%"+*where.Keyword+"%")
		}
		if where.SecretId != nil {
			query = query.Where("secret_id = ?", *where.SecretId)
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
