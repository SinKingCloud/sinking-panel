package software

import (
	repositorySoftware "server/app/repository/software"
	serviceSystem "server/app/service/system"
	"server/app/util/context"
	"server/app/util/page"
	"server/app/util/plugin"
	"sync"
	"sync/atomic"
)

// Service 软件包管理。软件实现从独立包加载，不再编译进面板。
type Service interface {
	Init(string, string, string) error
	Select(*repositorySoftware.SelectSoftware, *page.Query) (*page.Result[*repositorySoftware.Software], error)
	Info(int64) (*Info, error)
	Install(string) (*Info, error)
	Uninstall(string) (int64, error)
	Update(int64, string) error
	GetConfig(string) (string, error)
	SetConfig(string, string) error
	Invoke(*context.Context, string, string) error
	Page(string, string) ([]byte, string, string, error)
	Boot()
	Close()
}

type service struct {
	repository repositorySoftware.Interface
	system     serviceSystem.Service
	plugins    *plugin.Manager
	mu         sync.Mutex
	entries    map[int64]*entry
	importing  sync.Mutex
	closed     bool
	booted     atomic.Bool
}

// NewService 注入仓储、系统任务与已配置业务能力的插件管理器。
func NewService(repository repositorySoftware.Interface, system serviceSystem.Service, plugins *plugin.Manager) *service {
	return &service{
		repository: repository,
		system:     system,
		plugins:    plugins,
		entries:    make(map[int64]*entry),
	}
}
