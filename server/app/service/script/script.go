package script

import (
	"server/app/model"
	repositoryScript "server/app/repository/script"
	"server/app/util/page"
)

// Service service接口
type Service interface {
	Create(data *model.Script) error
	DeleteByIds(ids []int64) error
	FindById(id int64) (*model.Script, error)
	Select(where *repositoryScript.SelectScript, queryPage *page.Query) (*page.Result[*repositoryScript.Script], error)
	UpdateByIds(ids []int64, data *repositoryScript.UpdateScript) error
}

// service 注入结构
type service struct {
	repositoryScript repositoryScript.Interface
}

// NewService 实例化service
func NewService(repositoryScript repositoryScript.Interface) *service {
	return &service{repositoryScript: repositoryScript}
}
