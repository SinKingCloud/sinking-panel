package site

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"server/app/constant"
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
	if repositorySite == nil || repositorySiteDomain == nil || repositoryCert == nil || repositorySecret == nil || typeService == nil || configService == nil || database == nil || database.Db == nil || cache == nil {
		return nil, errors.New("网站服务依赖不能为空")
	}
	if len(options) > 1 {
		return nil, errors.New("网站服务参数最多只能传入一组")
	}
	config := Options{Root: constant.SitePath}
	if len(options) > 0 {
		config = options[0]
		if strings.TrimSpace(config.Root) == "" {
			config.Root = constant.SitePath
		}
	}
	root, err := filepath.Abs(config.Root)
	if err != nil {
		return nil, fmt.Errorf("解析网站服务目录失败: %w", err)
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("创建网站服务目录失败: %w", err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("解析网站服务目录失败: %w", err)
	}
	defaultHTTP := webServer.Options{}
	for _, item := range []struct {
		target *string
		path   string
	}{
		{&defaultHTTP.DataPath, constant.ServerDataPath},
		{&defaultHTTP.ACMEPath, constant.AcmePath},
		{&defaultHTTP.CachePath, constant.ServerCachePath},
		{&defaultHTTP.LogPath, constant.ServerLogPath},
		{&defaultHTTP.WAFLogPath, constant.ServerWAFLogPath},
		{&defaultHTTP.ConfigPath, constant.ServerConfigPath},
	} {
		*item.target, err = filepath.Abs(item.path)
		if err != nil {
			return nil, fmt.Errorf("解析 HTTP 服务默认路径失败: %w", err)
		}
		*item.target = filepath.Clean(*item.target)
	}
	result := &service{
		repositorySite:       repositorySite,
		repositorySiteDomain: repositorySiteDomain,
		repositoryCert:       repositoryCert,
		repositorySecret:     repositorySecret,
		typeService:          typeService,
		configService:        configService,
		cache:                cache,
		database:             database,
		root:                 filepath.Clean(root),
		active:               true,
	}
	result.process = processManager.NewManager()
	if len(options) == 0 {
		stored, exists, loadErr := result.loadHTTP()
		if loadErr != nil {
			log.Printf("网站 HTTP 配置损坏，已使用默认配置启动: %v", loadErr)
		} else if exists {
			config.HTTP = stored
		}
	}
	storedConfigs := configService.Group(constant.SiteGroup)
	if raw, exists := storedConfigs[constant.SiteHTTPEnabled]; exists {
		enabled, parseErr := strconv.ParseBool(strings.TrimSpace(raw))
		if parseErr != nil {
			log.Printf("网站 HTTP 服务状态配置损坏，已按启用状态处理: %v", parseErr)
		} else {
			result.active = enabled
		}
	}
	effectiveHTTP := config.HTTP
	effectiveHTTP.ACMEPath = defaultHTTP.ACMEPath
	for target, value := range map[*string]string{
		&effectiveHTTP.DataPath:   defaultHTTP.DataPath,
		&effectiveHTTP.CachePath:  defaultHTTP.CachePath,
		&effectiveHTTP.LogPath:    defaultHTTP.LogPath,
		&effectiveHTTP.WAFLogPath: defaultHTTP.WAFLogPath,
		&effectiveHTTP.ConfigPath: defaultHTTP.ConfigPath,
	} {
		if strings.TrimSpace(*target) == "" {
			*target = value
		}
	}
	httpManager, err := webServer.NewManager(root, effectiveHTTP)
	if err != nil {
		log.Printf("网站 HTTP 配置无法加载，已使用默认配置启动: %v", err)
		fallbackHTTP := defaultHTTP
		fallbackHTTP.ConfigPath = ""
		httpManager, err = webServer.NewManager(root, fallbackHTTP)
	}
	if err != nil {
		return nil, err
	}
	result.http = httpManager
	return result, nil
}
