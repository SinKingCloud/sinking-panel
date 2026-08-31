package log

import (
	"errors"
	"server/app/enum/log_type"
	"server/app/model"
	repositoryLog "server/app/repository/log"
	"server/app/util/page"
	"strconv"
)

// Select 获取数据
func (s *service) Select(where *repositoryLog.SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error) {
	if where != nil && where.Type != "" {
		value, err := strconv.Atoi(where.Type)
		if err != nil {
			return nil, errors.New("日志类型参数错误")
		}
		if _, ok := log_type.Map()[value]; !ok {
			return nil, errors.New("日志类型参数不合法")
		}
	}
	return s.repositoryLog.Select(where, queryPage)
}
