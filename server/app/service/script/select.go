package script

import (
	"errors"
	repositoryScript "server/app/repository/script"
	"server/app/util/page"
)

// Select 查询常用脚本
func (s *service) Select(where *repositoryScript.SelectScript, queryPage *page.Query) (*page.Result[*repositoryScript.Script], error) {
	if where != nil && where.TypeId != nil && *where.TypeId < 0 {
		return nil, errors.New("脚本分类参数错误")
	}
	return s.repositoryScript.Select(where, queryPage)
}
