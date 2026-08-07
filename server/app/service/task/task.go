package task

import (
	"server/app/model"
	repositoryTask "server/app/repository/task"
	"server/app/util/page"
	"sync"

	"github.com/robfig/cron/v3"
)

// Service service接口
type Service interface {
	Stop(id int64) error
	Restore(id int64) error
	Run(id int64) error
	Remove(ids []int64) error
	Refresh(id int64) error
	Add(data *model.Task) error
	Start()
	FindById(id int64) (*Info, error)
	Select(where *repositoryTask.SelectTask, queryPage *page.Query) (*page.Result[*repositoryTask.Task], error)
	ReadLog(id int64, cursor int64, before int64, pageSize int) map[string]interface{}
	ClearLog(id int64) error
	WriteLog(id int64, content string) error
	ValidateCron(expr string) bool
	UpdateByIds(ids []int64, data *repositoryTask.UpdateTask) error
}

// service 注入结构
type service struct {
	repositoryTask repositoryTask.Interface
	instance       *cron.Cron
	startOnce      sync.Once
	taskLock       sync.Mutex
	logLock        sync.RWMutex
}

// NewService 实例化service
func NewService(repository repositoryTask.Interface) *service {
	return &service{
		repositoryTask: repository,
		instance: cron.New(
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(cron.DiscardLogger),
			),
		),
	}
}
