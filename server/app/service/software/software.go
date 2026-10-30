package software

import (
	"server/app/model"
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
	Install(string) (int64, error)
	Uninstall(string) (int64, error)
	Update(int64, *repositorySoftware.UpdateSoftware) error
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
	manifests  map[string]*plugin.Manifest
	icons      map[string]string // 清单图标路径；与 entries 一样由 mu 保护
	cacheMu    sync.RWMutex
	records    map[string]*model.Software
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
		icons:      make(map[string]string),
		manifests:  make(map[string]*plugin.Manifest),
		records:    make(map[string]*model.Software),
	}
}
