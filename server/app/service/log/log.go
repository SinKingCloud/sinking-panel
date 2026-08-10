package log

import (
	"server/app/model"
	repositoryLog "server/app/repository/log"
	"server/app/util/page"
)

// Service service接口
type Service interface {
	Create(ip string, types int, title string, content string)
	Select(where *repositoryLog.SelectLog, queryPage *page.Query) (*page.Result[*model.Log], error)
}

// service 注入结构
type service struct {
	repositoryLog repositoryLog.Interface
}

// NewService 实例化service
func NewService(repositoryLog repositoryLog.Interface) *service {
	return &service{
		repositoryLog: repositoryLog,
	}
}
