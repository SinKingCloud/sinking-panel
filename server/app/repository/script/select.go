package script

import (
	"server/app/model"
	"server/app/util/page"
	"strconv"
	"strings"
)

// Select 查询常用脚本
func (r *Repository) Select(where *SelectScript, queryPage *page.Query) (*page.Result[*Script], error) {
	query := r.Database.Db.Model(&model.Script{})
	if where != nil {
		if where.Module != nil {
			modules := []string{*where.Module}
			if index := strings.LastIndexByte(*where.Module, '_'); index > 0 {
				if _, err := strconv.ParseUint((*where.Module)[index+1:], 10, 64); err == nil {
					modules = append(modules, (*where.Module)[:index])
				}
			}
			// 精确匹配全局及当前实例，避免实例 1 命中实例 10 的脚本。
			query = query.Where("module IN ?", modules)
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
