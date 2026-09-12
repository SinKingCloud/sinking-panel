package task

import (
	"context"
	"server/app/model"
	repositoryTask "server/app/repository/task"
	serviceTypes "server/app/service/types"
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
	Close()
	FindById(id int64) (*Info, error)
	Select(where *repositoryTask.SelectTask, queryPage *page.Query) (*page.Result[*repositoryTask.Task], error)
	ReadLog(id int64, after int64, before int64, pageSize int) (map[string]interface{}, error)
	ClearLog(id int64) error
	WriteLog(id int64, content string) error
	UpdateByIds(ids []int64, data *repositoryTask.UpdateTask) error
}

// service 注入结构
type service struct {
	repositoryTask repositoryTask.Interface
	typeService    serviceTypes.Service
	instance       *cron.Cron
	ctx            context.Context
	cancel         context.CancelFunc
	startOnce      sync.Once
	closeOnce      sync.Once
	runWait        sync.WaitGroup
	closed         bool
	taskLock       sync.Mutex
	logLock        sync.RWMutex
}

// NewService 实例化service
func NewService(repositoryTask repositoryTask.Interface, typeService serviceTypes.Service) *service {
	ctx, cancel := context.WithCancel(context.Background())
	return &service{
		repositoryTask: repositoryTask,
		typeService:    typeService,
		ctx:            ctx,
		cancel:         cancel,
		instance: cron.New(
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(cron.DiscardLogger),
			),
		),
	}
}
