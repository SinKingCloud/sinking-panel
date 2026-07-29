package log

import (
	"server/app/model"
	repositoryLog "server/app/repository/log"
	"server/app/util/page"
)

// Select 获取数据
func (s *service) Select(where *repositoryLog.SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error) {
	return s.repositoryLog.Select(where, queryPage)
}
