package site

import (
	"context"
	"sync"

	"server/app/model"
	repositoryCert "server/app/repository/cert"
	repositorySecret "server/app/repository/secret"
	repositorySite "server/app/repository/site"
	repositorySiteDomain "server/app/repository/site_domain"
	serviceConfig "server/app/service/config"
	serviceTypes "server/app/service/types"
	"server/app/util/cache"
	"server/app/util/database"
	"server/app/util/page"
	processManager "server/app/util/process"
	webServer "server/app/util/server"
)

// Service 网站管理服务接口。
type Service interface {
	Boot() func()
	Start() error
	Stop() error
	Restart() error
	Running() bool
	Sync() error
	GetHTTP() webServer.Options
	UpdateHTTP(config *HTTPUpdate) error
	ReadServerLog(after int64, before int64, pageSize int) (map[string]interface{}, error)
	ClearServerLog() error
	ReadLog(id int64, logType webServer.LogType, after int64, before int64, pageSize int) (map[string]interface{}, error)
	ClearLog(id int64, logType webServer.LogType) error
	Create(data *CreateSite) (*Site, error)
	Update(id int64, data *UpdateSite) error
	GetDomains(id int64) ([]*model.SiteDomain, error)
	UpdateDomains(id int64, domains []string) error
	GetSSL(id int64) (*SSLSettings, error)
	UpdateSSL(id int64, config *TLSUpdate, bindings []SiteDomainCertificate) error
	GetWAF(id int64) (*webServer.WAFOptions, error)
	UpdateWAF(id int64, config *WAFUpdate) error
	GetCache(id int64) (*CacheConfig, error)
	UpdateCache(id int64, config *CacheUpdate) error
	GetRateLimit(id int64) (*webServer.RateLimitOptions, error)
	UpdateRateLimit(id int64, config *RateLimitUpdate) error
	GetTrafficLimit(id int64) (*webServer.TrafficLimitOptions, error)
	UpdateTrafficLimit(id int64, config *TrafficLimitUpdate) error
	GetHeaders(id int64) (*webServer.HeaderOptions, error)
	UpdateHeaders(id int64, config *HeaderUpdate) error
	GetCompression(id int64) (*webServer.CompressionOptions, error)
	UpdateCompression(id int64, config *CompressionUpdate) error
	GetRedirects(id int64) ([]RedirectConfig, error)
	UpdateRedirects(id int64, redirects []RedirectConfig) error
	GetRoutes(id int64) ([]RouteConfig, error)
	UpdateRoutes(id int64, routes []RouteConfig) error
	GetStatic(id int64) (*StaticOptions, error)
	UpdateStatic(id int64, config *StaticUpdate) error
	GetProxy(id int64) (*webServer.ProxyOptions, error)
	UpdateProxy(id int64, config *ProxyUpdate) error
	GetFastCGI(id int64) (*webServer.ProxyOptions, error)
	UpdateFastCGI(id int64, config *FastCGIUpdate) error
	GetProcess(id int64) (*ProcessConfig, error)
	UpdateProcess(id int64, config *ProcessUpdate) error
	Process(id int64, action string) (*processManager.Status, error)
	Delete(id int64, deleteRoot bool) error
	Enable(id int64) error
	Disable(id int64) error
	FindById(id int64) (*Site, error)
	GetIdNameMap(refresh bool) (map[int64]string, error)
	Select(where *repositorySite.SelectSite, queryPage *page.Query) (*page.Result[*repositorySite.Site], error)
	ClearCache(id int64) error
	CreateCert(data *model.Cert) error
	UpdateCert(id int64, data *repositoryCert.UpdateCert) error
	DeleteCert(id int64) error
	FindCert(id int64) (*model.Cert, error)
	SelectCert(where *repositoryCert.SelectCert, queryPage *page.Query) (*page.Result[*repositoryCert.Cert], error)
	ObtainCert(ctx context.Context, name string, request CertificateRequest) (*model.Cert, error)
	RenewCert(ctx context.Context, id int64, request CertificateRequest) (*model.Cert, error)
}

// service 注入网站、域名、证书仓储和运行时管理器。
type service struct {
	repositorySite       repositorySite.Interface
	repositorySiteDomain repositorySiteDomain.Interface
	repositoryCert       repositoryCert.Interface
	repositorySecret     repositorySecret.Interface
	typeService          serviceTypes.Service
	configService        serviceConfig.Service
	cache                cache.Interface
	database             *database.Database
	http                 *webServer.Manager
	process              *processManager.Manager
	processes            []processManager.Config
	root                 string
	active               bool
	operationMu          sync.Mutex
	logMu                sync.RWMutex
}

// NewService 创建网站管理服务。
func NewService(repositorySite repositorySite.Interface, repositorySiteDomain repositorySiteDomain.Interface, repositoryCert repositoryCert.Interface, repositorySecret repositorySecret.Interface, typeService serviceTypes.Service, configService serviceConfig.Service, database *database.Database, cache cache.Interface, options ...Options) (*service, error) {
	return newService(repositorySite, repositorySiteDomain, repositoryCert, repositorySecret, typeService, configService, database, cache, options...)
}
