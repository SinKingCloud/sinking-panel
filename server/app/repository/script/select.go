package script

import (
	"server/app/model"
	"server/app/util/page"
	"strings"
)

// Select 查询常用脚本
func (r *Repository) Select(where *SelectScript, queryPage *page.Query) (*page.Result[*Script], error) {
	query := r.Database.Db.Model(&model.Script{})
	if where != nil {
		if where.Module != nil {
			module, _, instance := strings.Cut(*where.Module, "_")
			if instance {
				// 精确匹配全局及当前实例，避免 ssh_1 命中 ssh_10 等其他实例。
				query = query.Where("module IN ?", []string{module, *where.Module})
			} else {
				query = query.Where("module = ?", *where.Module)
			}
		}
		if where.TypeId != nil {
			query = query.Where("type_id = ?", *where.TypeId)
		}
		if where.Keyword != nil {
			keyword := "%" + *where.Keyword + "%"
			query = query.Where("(name LIKE ? OR script LIKE ?)", keyword, keyword)
		}
		if where.Name != nil {
			query = query.Where("name LIKE ?", "%"+*where.Name+"%")
		}
		if where.Script != nil {
			query = query.Where("script LIKE ?", "%"+*where.Script+"%")
		}
	}
	return r.Repository.SelectPage(query, queryPage)
}
