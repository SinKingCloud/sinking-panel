package script

import (
	"server/app/model"
	repositoryScript "server/app/repository/script"
	"server/app/util/page"
)

// Select 查询常用脚本
func (s *service) Select(where *repositoryScript.SelectScript, queryPage *page.Query) (*page.Result[*model.Script], error) {
	return s.repositoryScript.Select(where, queryPage)
}
