package task

import (
	"context"
	"server/app/constant"
	"server/app/model"
	repositorySite "server/app/repository/site"
	repositoryTask "server/app/repository/task"
	serviceOss "server/app/service/oss"
	"server/app/service/task/backup"
	"server/app/service/task/request"
	"server/app/service/task/script"
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
	Cancel(ids []int64) error // 批量取消并等待任务结束，保留数据库配置和日志
	Remove(ids []int64) error
	Refresh(id int64) error // 按保存的配置重新挂载任务
	Add(data *model.Task, run ...RunFunc) error
	Start()
	Close()
	FindById(id int64) (*Info, error)
	Select(where *repositoryTask.SelectTask, queryPage *page.Query) (*page.Result[*repositoryTask.Task], error)
	ReadLog(id int64, after int64, before int64, pageSize int) (map[string]interface{}, error)
	ClearLog(id int64) error
	UpdateByIds(ids []int64, data *repositoryTask.UpdateTask) error
}

// service 注入结构
type service struct {
	repositoryTask repositoryTask.Interface
	typeService    serviceTypes.Service
	instance       *cron.Cron
	internalTasks  map[int64]*job
	jobs           map[*job]struct{}
	ctx            context.Context
	cancel         context.CancelFunc
	startOnce      sync.Once
	closeOnce      sync.Once
	runWait        sync.WaitGroup
	closed         bool
	taskLock       sync.Mutex
	logLock        sync.RWMutex
	script         *script.Manager
	request        *request.Manager
	backup         *backup.Manager
}

// NewService 创建调度服务，网站备份通过仓储读取数据，不依赖网站运行服务。
func NewService(repositoryTask repositoryTask.Interface, repositorySite repositorySite.Interface, typeService serviceTypes.Service, ossService serviceOss.Service) *service {
	ctx, cancel := context.WithCancel(context.Background())
	return &service{
		repositoryTask: repositoryTask,
		typeService:    typeService,
		script:         script.NewManager(),
		request:        request.NewManager(),
		backup:         backup.NewManager(repositorySite, ossService, constant.BackupPath),
		internalTasks:  make(map[int64]*job),
		jobs:           make(map[*job]struct{}),
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
